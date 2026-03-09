import React from 'react';
import { useState } from 'react';
import './Login.css'; // External CSS for styling

const Login = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');

    const handleSubmit = (e) => {
        e.preventDefault();
        // Handle login
    };

    const handleGoogleSignIn = () => {
        // Handle Google sign-in
    };

    const handleGitHubSignIn = () => {
        // Handle GitHub sign-in
    };

    return (
        <div className='login-container'>
            <h2>Login</h2>
            <form onSubmit={handleSubmit} className='login-form'>
                <input
                    type='email'
                    placeholder='Email'
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                />
                <input
                    type='password'
                    placeholder='Password'
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                />
                <button type='submit'>Login</button>
            </form>
            <div className='social-signin'>
                <button onClick={handleGoogleSignIn}>Sign in with Google</button>
                <button onClick={handleGitHubSignIn}>Sign in with GitHub</button>
            </div>
        </div>
    );
};

export default Login;
