package app

import (
	"fmt"
	"github.com/OB1Company/filehive/fil"
	"github.com/OB1Company/filehive/repo"
	"github.com/OB1Company/filehive/repo/models"
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"gorm.io/gorm"
	"io"
	"math/big"
	"net/http"
	"os"
	"path"
	"testing"
	"time"
)

func Test_Handlers(t *testing.T) {
	t.Run("User Tests", func(t *testing.T) {
		runAPITests(t, apiTests{
			{
				name:             "Post user success",
				path:             "/api/v1/user",
				method:           http.MethodPost,
				statusCode:       http.StatusOK,
				body:             []byte(`{"email": "brian@ob1.io", "password":"asdf", "name": "Brian", "country": "United_States"}`),
				expectedResponse: nil,
			},
			{
				name:             "Post user invalid JSON",
				path:             "/api/v1/user",
				method:           http.MethodPost,
				statusCode:       http.StatusBadRequest,
				body:             []byte(`{"email": "brian@ob1.io "password":"asdf", "name": "Brian", "country": "United_States"}`),
				expectedResponse: errorReturn(ErrInvalidJSON),
			},
			{
				name:             "Post user nil password",
				path:             "/api/v1/user",
				method:           http.MethodPost,
				statusCode:       http.StatusBadRequest,
				body:             []byte(`{"email": "brian2@ob1.io", "password":"", "name": "Brian", "country": "United_States"}`),
				expectedResponse: errorReturn(ErrBadPassword),
			},
			{
				name:             "Post user invalid email",
				path:             "/api/v1/user",
				method:           http.MethodPost,
				statusCode:       http.StatusBadRequest,
				body:             []byte(`{"email": "brian2ob1", "password":"adsf", "name": "Brian", "country": "United_States"}`),
				expectedResponse: errorReturn(ErrInvalidEmail),
			},
			{
				name:             "Post user already exists",
				path:             "/api/v1/user",
				method:           http.MethodPost,
				statusCode:       http.StatusConflict,
				body:             []byte(`{"email": "brian@ob1.io", "password":"", "name": "Brian", "country": "United_States"}`),
				expectedResponse: errorReturn(ErrUserExists),
			},
			{
				name:       "Get user while logged in",
				path:       "/api/v1/user",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Email   string
					Name    string
					Country string
					Avatar  string
				}{
					Email:   "brian@ob1.io",
					Name:    "Brian",
					Country: "United_States",
				}),
			},
			{
				name:       "Get user from path",
				path:       "/api/v1/user/brian@ob1.io",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Email   string
					Name    string
					Country string
					Avatar  string
				}{
					Email:   "brian@ob1.io",
					Name:    "Brian",
					Country: "United_States",
				}),
			},
			{
				name:             "Get user from path not found",
				path:             "/api/v1/user/chris@ob1.io",
				method:           http.MethodGet,
				statusCode:       http.StatusNotFound,
				expectedResponse: errorReturn(ErrUserNotFound),
			},
			{
				name:             "Patch user success",
				path:             "/api/v1/user",
				method:           http.MethodPatch,
				statusCode:       http.StatusOK,
				body:             []byte(`{"email": "brian@ob1.io", "password":"ffff", "name": "Brian2", "country": "Botswana"}`),
				expectedResponse: nil,
			},
			{
				name:       "Check user patched correctly",
				path:       "/api/v1/user/brian@ob1.io",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Email   string
					Name    string
					Country string
					Avatar  string
				}{
					Email:   "brian@ob1.io",
					Name:    "Brian2",
					Country: "Botswana",
				}),
			},
			{
				name:             "Patch user change email",
				path:             "/api/v1/user",
				method:           http.MethodPatch,
				statusCode:       http.StatusOK,
				body:             []byte(`{"email": "brian2@ob1.io"}`),
				expectedResponse: nil,
			},
			{
				name:       "Check user patched correctly",
				path:       "/api/v1/user/brian2@ob1.io",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Email   string
					Name    string
					Country string
					Avatar  string
				}{
					Email:   "brian2@ob1.io",
					Name:    "Brian2",
					Country: "Botswana",
				}),
			},
			{
				name:             "Check previous email deleted correctly",
				path:             "/api/v1/user/brian@ob1.io",
				method:           http.MethodGet,
				statusCode:       http.StatusNotFound,
				expectedResponse: errorReturn(ErrUserNotFound),
			},
		})
	})

	t.Run("Login Tests", func(t *testing.T) {
		runAPITests(t, apiTests{
			{
				name:             "Post extend token not logged in",
				path:             "/api/v1/token/extend",
				method:           http.MethodPost,
				statusCode:       http.StatusUnauthorized,
				body:             nil,
				expectedResponse: errorReturn(ErrNotLoggedIn),
			},
			{
				name:             "Post login invalid email",
				path:             "/api/v1/login",
				method:           http.MethodPost,
				statusCode:       http.StatusUnauthorized,
				body:             []byte(`{"email": "brian@ob1.io", "password":"asdf"}`),
				expectedResponse: errorReturn(ErrIncorrectPassword),
			},
			{
				name:             "Post login invalid JSON",
				path:             "/api/v1/login",
				method:           http.MethodPost,
				statusCode:       http.StatusBadRequest,
				body:             []byte(`{"email": "brian@ob1.io", "password":"asdf"`),
				expectedResponse: errorReturn(ErrInvalidJSON),
			},
			{
				name:       "Post login incorrect password",
				path:       "/api/v1/login",
				method:     http.MethodPost,
				statusCode: http.StatusUnauthorized,
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					return db.Update(func(tx *gorm.DB) error {
						salt := []byte("salt")
						pw := hashPassword([]byte("asdf"), salt)
						return tx.Save(&models.User{
							Email:          "brian@ob1.io",
							Country:        "United_States",
							Name:           "Brian",
							Salt:           salt,
							HashedPassword: pw,
						}).Error
					})
				},
				body:             []byte(`{"email": "brian@ob1.io", "password":"aaaaa"}`),
				expectedResponse: errorReturn(ErrIncorrectPassword),
			},
			{
				name:             "Post login valid",
				path:             "/api/v1/login",
				method:           http.MethodPost,
				statusCode:       http.StatusOK,
				body:             []byte(`{"email": "brian@ob1.io", "password":"asdf"}`),
				expectedResponse: nil,
			},
			{
				name:       "Get user while logged in",
				path:       "/api/v1/user",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Email   string
					Name    string
					Country string
					Avatar  string
				}{
					Email:   "brian@ob1.io",
					Name:    "Brian",
					Country: "United_States",
				}),
			},
			{
				name:             "Post extend token",
				path:             "/api/v1/token/extend",
				method:           http.MethodPost,
				statusCode:       http.StatusOK,
				body:             nil,
				expectedResponse: nil,
			},
		})
	})

	t.Run("Image Tests", func(t *testing.T) {
		runAPITests(t, apiTests{
			{
				name:             "Post user success",
				path:             "/api/v1/user",
				method:           http.MethodPost,
				statusCode:       http.StatusOK,
				body:             []byte(`{"email": "brian@ob1.io", "password":"asdf", "name": "Brian", "country": "United_States"}`),
				expectedResponse: nil,
			},
			{
				name:             "Patch user with avatar",
				path:             "/api/v1/user",
				method:           http.MethodPatch,
				statusCode:       http.StatusOK,
				body:             []byte(fmt.Sprintf(`{"avatar": "%s"}`, jpgTestImage)),
				expectedResponse: nil,
			},
			{
				name: "Get avatar",
				path: "/api/v1/image/avatar-1.jpg",
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					return db.View(func(db *gorm.DB) error {
						// Since we don't know the filename from prior API call we will look it
						// up in the db and create a new file with the name avatar-1.jpg so we
						// can test loading the avatar image.
						var user models.User
						err := db.Where("email=?", "brian@ob1.io").First(&user).Error
						if err != nil {
							return err
						}
						f1, err := os.Open(path.Join(testStaticDir, "images", user.AvatarFilename))
						if err != nil {
							return err
						}
						f2, err := os.Create(path.Join(testStaticDir, "images", "avatar-1.jpg"))
						if err != nil {
							return err
						}
						_, err = io.Copy(f2, f1)
						if err != nil {
							return err
						}
						return nil
					})
				},
				method:           http.MethodGet,
				statusCode:       http.StatusOK,
				expectedResponse: jpgImageBytes,
			},
		})
	})

	t.Run("Wallet Tests", func(t *testing.T) {
		runAPITests(t, apiTests{
			{
				name:       "Post user success",
				path:       "/api/v1/user",
				method:     http.MethodPost,
				statusCode: http.StatusOK,
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					addr, err := address.NewFromString("f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi")
					if err != nil {
						return err
					}
					wbe.(*fil.MockWalletBackend).SetNextAddress(addr)
					return nil
				},
				body:             []byte(`{"email": "brian@ob1.io", "password":"asdf", "name": "Brian", "country": "United_States"}`),
				expectedResponse: nil,
			},
			{
				name:       "Get wallet address",
				path:       "/api/v1/wallet/address",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Address string
				}{
					Address: "f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi",
				}),
			},
			{
				name:       "Get wallet balance",
				path:       "/api/v1/wallet/balance",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					wbe.(*fil.MockWalletBackend).SetNextTime(time.Time{})
					txid, err := cid.Decode("bafkreiewgqfti56ls5zt2kko2utajoliipl3te7cl5lvtiowgny6qb2pde")
					if err != nil {
						return err
					}
					wbe.(*fil.MockWalletBackend).SetNextTxid(txid)
					addr, err := address.NewFromString("f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi")
					if err != nil {
						return err
					}
					amt, _ := new(big.Int).SetString("15500000000000000000", 10)
					wbe.(*fil.MockWalletBackend).GenerateToAddress(addr, amt)
					return nil
				},
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Balance float64
				}{
					Balance: 15.5,
				}),
			},
			{
				name:       "Post wallet send",
				path:       "/api/v1/wallet/send",
				method:     http.MethodPost,
				statusCode: http.StatusOK,
				body:       []byte(`{"address": "f1gyvikksfdmokwhg5jhcrkvfqkyd2sjdy46klgbq", "amount": 1}`),
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					txid, err := cid.Decode("bafkreif2mzhq6663465bcb2s3xgqefysbmr3a2bxloobw7s4vrxooj6kva")
					if err != nil {
						return err
					}
					wbe.(*fil.MockWalletBackend).SetNextTxid(txid)
					return nil
				},
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Txid string
				}{
					Txid: "bafkreif2mzhq6663465bcb2s3xgqefysbmr3a2bxloobw7s4vrxooj6kva",
				}),
			},
			{
				name:             "Post wallet send insufficient funds",
				path:             "/api/v1/wallet/send",
				method:           http.MethodPost,
				statusCode:       http.StatusBadRequest,
				body:             []byte(`{"address": "f1gyvikksfdmokwhg5jhcrkvfqkyd2sjdy46klgbq", "amount": 20}`),
				expectedResponse: errorReturn(fil.ErrInsuffientFunds),
			},
			{
				name:       "Get wallet balance",
				path:       "/api/v1/wallet/balance",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON(struct {
					Balance float64
				}{
					Balance: 14.5,
				}),
			},
			{
				name:       "Get wallet transactions",
				path:       "/api/v1/wallet/transactions",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON([]struct {
					To            string  `json:"to"`
					From          string  `json:"from"`
					TransactionID string  `json:"transactionID"`
					Amount        float64 `json:"amount"`
					Timestamp     string  `json:"timestamp"`
				}{
					{
						Timestamp:     "0001-01-01T00:00:00Z",
						Amount:        15.5,
						To:            "f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi",
						From:          "",
						TransactionID: "bafkreiewgqfti56ls5zt2kko2utajoliipl3te7cl5lvtiowgny6qb2pde",
					},
					{
						Timestamp:     "0001-01-01T00:00:00Z",
						Amount:        1,
						To:            "f1gyvikksfdmokwhg5jhcrkvfqkyd2sjdy46klgbq",
						From:          "f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi",
						TransactionID: "bafkreif2mzhq6663465bcb2s3xgqefysbmr3a2bxloobw7s4vrxooj6kva",
					},
				}),
			},
			{
				name:       "Get wallet transactions with limit",
				path:       "/api/v1/wallet/transactions?limit=1",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON([]struct {
					To            string  `json:"to"`
					From          string  `json:"from"`
					TransactionID string  `json:"transactionID"`
					Amount        float64 `json:"amount"`
					Timestamp     string  `json:"timestamp"`
				}{
					{
						Timestamp:     "0001-01-01T00:00:00Z",
						Amount:        15.5,
						To:            "f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi",
						From:          "",
						TransactionID: "bafkreiewgqfti56ls5zt2kko2utajoliipl3te7cl5lvtiowgny6qb2pde",
					},
				}),
			},
			{
				name:       "Get wallet transactions with offset",
				path:       "/api/v1/wallet/transactions?offset=1",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				expectedResponse: mustMarshalAndSanitizeJSON([]struct {
					To            string  `json:"to"`
					From          string  `json:"from"`
					TransactionID string  `json:"transactionID"`
					Amount        float64 `json:"amount"`
					Timestamp     string  `json:"timestamp"`
				}{
					{
						Timestamp:     "0001-01-01T00:00:00Z",
						Amount:        1,
						To:            "f1gyvikksfdmokwhg5jhcrkvfqkyd2sjdy46klgbq",
						From:          "f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi",
						TransactionID: "bafkreif2mzhq6663465bcb2s3xgqefysbmr3a2bxloobw7s4vrxooj6kva",
					},
				}),
			},
			{
				name:             "Get wallet transactions invalid limit",
				path:             "/api/v1/wallet/transactions?limit=zzz",
				method:           http.MethodGet,
				statusCode:       http.StatusBadRequest,
				expectedResponse: errorReturn(ErrInvalidOption),
			},
			{
				name:             "Get wallet transactions invalid offset",
				path:             "/api/v1/wallet/transactions?offset=zzz",
				method:           http.MethodGet,
				statusCode:       http.StatusBadRequest,
				expectedResponse: errorReturn(ErrInvalidOption),
			},
		})
	})

	t.Run("Dataset Tests", func(t *testing.T) {
		runAPITests(t, apiTests{
			{
				name:       "Post user success",
				path:       "/api/v1/user",
				method:     http.MethodPost,
				statusCode: http.StatusOK,
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					addr, err := address.NewFromString("f1cu3c2dqsbyt7nq63x2yubyy6ofuini2nfvnnahi")
					if err != nil {
						return err
					}
					wbe.(*fil.MockWalletBackend).SetNextAddress(addr)
					return nil
				},
				body:             []byte(`{"email": "brian@ob1.io", "password":"asdf", "name": "Brian", "country": "United_States"}`),
				expectedResponse: nil,
			},
			{
				name:        "Post dataset success",
				path:        "/api/v1/dataset",
				method:      http.MethodPost,
				statusCode:  http.StatusOK,
				contentType: "multipart/form-data; boundary=cc0ce5746707c1948657e8d0a2ca5570c2ddfd90ae6b7d5b49eac967c527",
				body: []byte(`--cc0ce5746707c1948657e8d0a2ca5570c2ddfd90ae6b7d5b49eac967c527
Content-Disposition: form-data; name="metadata"
Content-Type: application/json

{"title":"Snowden Leaks", "shortDescription": "This is a short description", "fullDescription": "This is a long description", "fileType": ".txt", "price": 1.234, "image": "/9j/4AAQSkZJRgABAQAAAQABAAD//gA7Q1JFQVRPUjogZ2QtanBlZyB2MS4wICh1c2luZyBJSkcgSlBFRyB2NjIpLCBxdWFsaXR5ID0gNjUK/9sAQwALCAgKCAcLCgkKDQwLDREcEhEPDxEiGRoUHCkkKyooJCcnLTJANy0wPTAnJzhMOT1DRUhJSCs2T1VORlRAR0hF/9sAQwEMDQ0RDxEhEhIhRS4nLkVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVF/8AAEQgAMgAyAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//EALUQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+v/EAB8BAAMBAQEBAQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAECdwABAgMRBAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm5+jp6vLz9PX29/j5+v/aAAwDAQACEQMRAD8A840awhv5zFKWDYyMHrVvWtE/szynj3GJ+MnsaoWFw1ndxTr1Rskeor0+70uPXNBYQ4JkQSRH36iiXw3CO9meWxxNJIqICWY4AHeu5g8C232aMztL5pUFtpGM/lUXgPw+13qD3lwhEdscAEdX/wDrVseNddl0l4bSxcLcN8zHAOB6c1UnyJLqxRTlJ9kY83guzQcNN/30P8KwNY0W206AvufceFBPWvRtMtrw6RHLqUm+dxvOVA2j04rzjxJqAv8AUXEZ/cxHavv71M20+UcbNc3Q5/bRUu2igCVRXpfw51MXFtJp0rfPD88ee6nr+R/nXmq13fw40xpL6TUXyEiGxfcnrVwV7kSdrHo7C10ixnn2rFEu6R8cZPU15r4espfFviua/uQTbxvvbPT/AGVrX+IetMyQ6PakmSUhpAv6Cuh0DT4PC/hsGbCsE82ZvfHSs4O160umiLmtFTW73/rzMjx7rC6Zp32WFgJ7gY4/hXua8nbmtTXtWk1rVZruQnDHCL/dXtWW1TBPd7suVl7q6EeKKKKsgkt42nlSNBlmIAr1rTZINA0MAkBYU3MfU15joEsEN5588irs+6GPetfX9cW8iis7eVSjHLsDxRJ+7yx3YRV5XeyNjwlbPrviCbWL0bkjbcoPQt2H4Vf+IOsTyxpplrHIyt80rKpwfQU3R9W0vS9PitkvIBtHzHeOT3q+3ibTiP8Aj9g/77FE+R2itkEXK7m92eXm2uB1gk/74NRPDKoOY3H1U16RceI7FgcXkJ/4GKwdT1q2lgkVJ0YlSOGoco20CzONzRUe6igBq1KtFFAD6KKKAGmo26UUUAR0UUUAf//Z"}
--cc0ce5746707c1948657e8d0a2ca5570c2ddfd90ae6b7d5b49eac967c527
Content-Disposition: form-data; name="file"; filename="snowden.txt"
Content-Type: application/octet-stream

Snowden Files

--cc0ce5746707c1948657e8d0a2ca5570c2ddfd90ae6b7d5b49eac967c527--`),
				expectedResponse: nil,
			},
			{
				name:       "Patch dataset",
				path:       "/api/v1/dataset",
				method:     http.MethodPatch,
				statusCode: http.StatusOK,
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					return db.Update(func(db *gorm.DB) error {
						var dataset models.Dataset
						err := db.Where("title=?", "Snowden Leaks").First(&dataset).Error
						if err != nil {
							return err
						}
						if err := db.Delete(&dataset).Error; err != nil {
							return err
						}
						dataset.ID = "1234"
						return db.Save(&dataset).Error
					})
				},
				body:             []byte(`{"title": "Changed title", "id": "1234"}`),
				expectedResponse: nil,
			},
			{
				name:       "Patch dataset unauthorized user",
				path:       "/api/v1/dataset",
				method:     http.MethodPatch,
				statusCode: http.StatusUnauthorized,
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					return db.Update(func(db *gorm.DB) error {
						var dataset models.Dataset
						err := db.Where("title=?", "Changed title").First(&dataset).Error
						if err != nil {
							return err
						}
						dataset.UserID = "ABCD"
						return db.Save(&dataset).Error
					})
				},
				body:             []byte(`{"title": "Changed title", "id": "1234"}`),
				expectedResponse: nil,
			},
			{
				name:             "Patch dataset not found",
				path:             "/api/v1/dataset",
				method:           http.MethodPatch,
				statusCode:       http.StatusBadRequest,
				body:             []byte(`{"title": "Changed title", "id": "1111"}`),
				expectedResponse: nil,
			},
			{
				name:       "Get dataset",
				path:       "/api/v1/dataset/1234",
				method:     http.MethodGet,
				statusCode: http.StatusOK,
				setup: func(db *repo.Database, wbe fil.WalletBackend) error {
					return db.Update(func(db *gorm.DB) error {
						var dataset models.Dataset
						err := db.Where("title=?", "Changed title").First(&dataset).Error
						if err != nil {
							return err
						}
						dataset.ImageFilename = "1AYAVn7Jq2UXcpMnHFqE4YMoLY1S2oUjyrkbPGHU88ndZg.jpg"
						dataset.JobID = "bafkreibsth7fjp4n45bvrrcn7edtx6jz7b6ghasce4stxg3u4olhqsfb7y"
						return db.Save(&dataset).Error
					})
				},
				expectedResponse: mustMarshalAndSanitizeJSON(models.Dataset{
					JobID:            "bafkreibsth7fjp4n45bvrrcn7edtx6jz7b6ghasce4stxg3u4olhqsfb7y",
					Price:            0,
					UserID:           "ABCD",
					FileType:         ".txt",
					Title:            "Changed title",
					ShortDescription: "This is a short description",
					FullDescription:  "This is a long description",
					ImageFilename:    "1AYAVn7Jq2UXcpMnHFqE4YMoLY1S2oUjyrkbPGHU88ndZg.jpg",
					ID:               "1234",
				}),
			},
			{
				name:             "Get dataset not found",
				path:             "/api/v1/dataset/4567",
				method:           http.MethodGet,
				statusCode:       http.StatusNotFound,
				expectedResponse: errorReturn(ErrDatasetNotFound),
			},
		})
	})
}
