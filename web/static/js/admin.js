document.addEventListener('DOMContentLoaded', () => {
    const logoutBtn = document.getElementById('logoutBtn');
    const createUserForm = document.getElementById('createUserForm');
    const usersTableBody = document.querySelector('#usersTable tbody');
    const createUserMessage = document.getElementById('createUserMessage');

    const getJwtToken = () => localStorage.getItem('jwt_token');

    // Redirect to login if no token
    if (!getJwtToken()) {
        window.location.href = '/web/login';
        return;
    }

    // Logout functionality
    logoutBtn.addEventListener('click', () => {
        localStorage.removeItem('jwt_token');
        window.location.href = '/web/login';
    });

    // Function to fetch and display users
    const fetchUsers = async () => {
        usersTableBody.innerHTML = '<tr><td colspan="4">Loading users...</td></tr>';
        try {
            const response = await fetch('/admin/users', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${getJwtToken()}`,
                    'Content-Type': 'application/json'
                }
            });

            if (response.status === 401 || response.status === 403) {
                // Token expired or not authorized, redirect to login
                localStorage.removeItem('jwt_token');
                window.location.href = '/web/login';
                return;
            }

            const data = await response.json();

            if (response.ok) {
                usersTableBody.innerHTML = ''; // Clear existing rows
                data.users.forEach(user => {
                    const row = usersTableBody.insertRow();
                    row.insertCell().textContent = user.id;
                    row.insertCell().textContent = user.username;
                    row.insertCell().textContent = user.is_admin ? 'Yes' : 'No';
                    row.insertCell().textContent = new Date(user.created_at).toLocaleString();
                });
            } else {
                usersTableBody.innerHTML = `<tr><td colspan="4" style="color:red;">Error loading users: ${data.error || 'Unknown error'}</td></tr>`;
            }
        } catch (error) {
            usersTableBody.innerHTML = `<tr><td colspan="4" style="color:red;">Network error: ${error.message}</td></tr>`;
        }
    };

    // Create User functionality
    createUserForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        const newUsername = document.getElementById('newUsername').value;
        const newPassword = document.getElementById('newPassword').value;
        const isAdmin = document.getElementById('isAdmin').checked;

        createUserMessage.textContent = ''; // Clear previous messages

        try {
            const response = await fetch('/admin/users', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${getJwtToken()}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username: newUsername, password: newPassword, is_admin: isAdmin })
            });

            if (response.status === 401 || response.status === 403) {
                localStorage.removeItem('jwt_token');
                window.location.href = '/web/login';
                return;
            }

            const data = await response.json();

            if (response.ok) {
                createUserMessage.className = 'message success';
                createUserMessage.textContent = `User ${data.username} created. API Key: ${data.api_key}, API Secret: ${data.api_secret}`;
                createUserForm.reset(); // Clear form
                fetchUsers(); // Refresh user list
            } else {
                createUserMessage.className = 'message error';
                createUserMessage.textContent = data.error || 'Failed to create user.';
            }
        } catch (error) {
            createUserMessage.className = 'message error';
            createUserMessage.textContent = 'Network error: ' + error.message;
        }
    });

    // Initial fetch of users
    fetchUsers();
});