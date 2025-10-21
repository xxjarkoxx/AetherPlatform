import * as React from 'react';
import { useState } from 'react';
import { useAuth } from '../context/Authcontext';
import { useNavigate } from 'react-router-dom';
import { FiUser, FiLock } from 'react-icons/fi';
import { FaGoogle, FaMicrosoft, FaGithub } from 'react-icons/fa';
import './login.css'; 

export const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();
    const handleSocialLogin = (provider: string) => {
    window.location.href = `/auth/oidc/${provider}`; // Ajusta la ruta según tu backend
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      await login(email, password);
      navigate('/dashboard');
    } catch (err) {
      setError('Credenciales incorrectas. Por favor, inténtalo de nuevo.');
      console.error(err);
    }
  };

  return (
    <div className="login-container">
      <div className="login-box">
        <h1>AetherPlatform</h1>
        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <label htmlFor="email">Email</label>
            <div className="input-icon-wrapper">
              <FiUser className="input-icon" />
              <input
                type="email"
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                placeholder="Your email"
              />
            </div>
          </div>
          <div className="input-group">
            <label htmlFor="password">Password</label>
            <div className="input-icon-wrapper">
              <FiLock className="input-icon" />
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                placeholder="Your password"
              />
            </div>
          </div>
          {error && <p className="error-message">{error}</p>}
          <button type="submit" className="login-button">
            Sign in
          </button>
        </form>
        <div className="divider">o</div>
        <div className="social-login-group">
          <button
            className="social-login google"
            onClick={() => handleSocialLogin('google')}
            type="button"
          >
            <FaGoogle className="social-icon" />
            Sign in with Google
          </button>
          <button
            className="social-login microsoft"
            onClick={() => handleSocialLogin('microsoft')}
            type="button"
          >
            <FaMicrosoft className="social-icon" />
            Sign in with Microsoft
          </button>
          <button
            className="social-login github"
            onClick={() => handleSocialLogin('github')}
            type="button"
          >
            <FaGithub className="social-icon" />
            Sign in with GitHub
          </button>
        </div>
      </div>
    </div>
  );
};