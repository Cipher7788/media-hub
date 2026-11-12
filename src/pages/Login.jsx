import React, { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { emailPasswordLogin, githubLogin, googleLogin, register } from '../services/auth';
import './Login.css';

function Login() {
  const { currentUser } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [creatingAccount, setCreatingAccount] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');

  if (currentUser) return <Navigate to="/" replace />;

  async function authenticate(action) {
    setPending(true);
    setError('');
    try {
      await action();
    } catch (err) {
      const messages = {
        'auth/invalid-credential': 'Check your email and password and try again.',
        'auth/email-already-in-use': 'This email already has an account. Sign in instead.',
        'auth/weak-password': 'Choose a stronger password.',
        'auth/popup-closed-by-user': 'The sign-in window was closed. Try again.',
        'auth/account-exists-with-different-credential': 'Sign in using the provider already linked to this email.',
        'auth/operation-not-allowed': 'This sign-in method is disabled.',
      };
      setError(messages[err.code] || 'Authentication failed. Please try again.');
    } finally {
      setPending(false);
    }
  }

  function handleSubmit(event) {
    event.preventDefault();
    const action = creatingAccount ? register : emailPasswordLogin;
    authenticate(() => action(email.trim(), password));
  }

  return (
    <div className="login-container">
      <h2>{creatingAccount ? 'Create account' : 'Login'}</h2>
      <form onSubmit={handleSubmit} className="login-form" aria-busy={pending}>
        <input
          type="email"
          aria-label="Email"
          placeholder="Email"
          autoComplete="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          disabled={pending}
          required
        />
        <input
          type="password"
          aria-label="Password"
          placeholder="Password"
          autoComplete={creatingAccount ? 'new-password' : 'current-password'}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          disabled={pending}
          required
        />
        <button type="submit" disabled={pending}>
          {creatingAccount ? 'Create account' : 'Sign in'}
        </button>
      </form>
      <button type="button" disabled={pending} onClick={() => {
        setCreatingAccount(!creatingAccount);
        setError('');
      }}>
        {creatingAccount ? 'Back to sign in' : 'Create an account'}
      </button>
      <div className="social-signin">
        <button type="button" disabled={pending} onClick={() => authenticate(googleLogin)}>
          Sign in with Google
        </button>
        <button type="button" disabled={pending} onClick={() => authenticate(githubLogin)}>
          Sign in with GitHub
        </button>
      </div>
      {pending && <p role="status">Connecting…</p>}
      {error && <p role="alert">{error}</p>}
    </div>
  );
}

export default Login;
