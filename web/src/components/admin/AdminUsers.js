import React, {useEffect, useState} from 'react'
import { Table, Thead, Tbody, Tr, Th, Td } from 'react-super-responsive-table';
import 'react-super-responsive-table/dist/SuperResponsiveTableStyle.css';
import {getAxiosInstance} from "../Auth";
import TimeAgo from "javascript-time-ago";

function UsersRows(props) {
    let rows = props.users.map((user)=> {
        return <UserRow key={user.id} metadata={user} selectHandler={props.selectHandler}/>;
    });

    return rows;
}

function UserRow(props) {
    const timeAgo = new TimeAgo('en-US');
    const user = props.metadata;
    const created = timeAgo.format(Date.parse(props.metadata.CreatedAt));

    return (
        <Tr>
            <Td><input type="checkbox" name="user" value={user.id} onChange={props.selectHandler}/></Td>
            <Td>{user.name}</Td>
            <Td>{user.email}</Td>
            <Td>{created} </Td>
            <Td>{user.admin ? "✅" : ""}</Td>
            <Td>{user.disabled ? "🚫" : "✅"}</Td>
            <Td>{user.powergateToken}</Td>
            <Td>{user.powergateID}</Td>
        </Tr>
    )
}

export default function AdminUsers() {

    const [users, setUsers] = useState([]);
    const [selectedCount, setSelectedCount] = useState(0);
    const [selectedUsers, setSelectedUsers] = useState([]);

    const refreshUsers = ()=>{
        const instance = getAxiosInstance();
        instance.get('/api/v1/users')
            .then((res)=>{
                const users = res.data.users;
                setUsers(users);

                const checkboxes = document.getElementsByName("user");
                for (let i = 0; i < checkboxes.length; i++) {
                    if (checkboxes[i].checked === true) {
                        checkboxes[i].checked = false;
                    }
                }

                setSelectedUsers([]);
                setSelectedCount(0);
            })
    }

    useEffect(() => {
        refreshUsers();
    }, []);

    const HandleDisable = (e)=>{
        e.preventDefault();
        const instance = getAxiosInstance();
        instance.post("/api/v1/users/disable", {"users":selectedUsers})
            .then((result)=>{
                refreshUsers();
            });

    }

    const HandleEnable = (e)=>{
        e.preventDefault();
        const instance = getAxiosInstance();
        instance.post("/api/v1/users/enable", {"users":selectedUsers})
            .then((result)=>{
                refreshUsers();
            });

    }

    const HandleAdmin = (e)=>{
        e.preventDefault();
        const instance = getAxiosInstance();
        instance.post("/api/v1/users/makeadmin", {"users":selectedUsers})
            .then((result)=>{
                refreshUsers();
            });

    }

    const HandleUser = (e)=>{
        e.preventDefault();
        const instance = getAxiosInstance();
        instance.post("/api/v1/users/makeuser", {"users":selectedUsers})
            .then((result)=>{
                refreshUsers();
            });

    }

    const HandleSelection = (e)=>{
        const checked = e.target.checked;
        if(checked) {
            setSelectedCount(selectedCount+1);
            let users = selectedUsers;
            users.push(e.target.value);
            setSelectedUsers(users);
        } else {
            setSelectedCount(selectedCount-1);
            let users = selectedUsers;
            const index = users.indexOf(e.target.value);
            if (index > -1) {
                users.splice(index, 1);
            }
            setSelectedUsers(users);
        }

    }

    return (
        <div className="margins-30 bottom-30">
            <h2>Users</h2>
            <br/>
            <div className="admin-toolbar">
                <div className="bold">{selectedCount} Selected</div>
                <div><a href="" className="orange-link2" onClick={HandleDisable}>Disable Account</a></div>
                <div><a href="" className="orange-link2" onClick={HandleEnable}>Enable Account</a></div>
                <div><a href="" className="orange-link2" onClick={HandleAdmin}>Make Admin</a></div>
                <div><a href="" className="orange-link2" onClick={HandleUser}>Make Normal</a></div>
            </div>
            <div>
                <Table className="sales-table font-12">
                    <Thead>
                        <Tr>
                            <Th></Th>
                            <Th>Name</Th>
                            <Th>Email</Th>
                            <Th>Created</Th>
                            <Th>Admin</Th>
                            <Th>Status</Th>
                            <Th>Token</Th>
                            <Th>ID</Th>
                        </Tr>
                    </Thead>
                    <Tbody>
                        <UsersRows users={users} selectHandler={HandleSelection}/>
                    </Tbody>
                </Table>
            </div>

        </div>
    );
}