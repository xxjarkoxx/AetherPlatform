import axios, { type AxiosInstance } from 'axios';

// Configuración base de la API
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

export interface User {
  id: string;
  email: string;
  name?: string;
}

class AuthService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Interceptor para agregar token a las peticiones
    this.api.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Interceptor para manejar respuestas de error
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Token expirado o inválido
          localStorage.removeItem('token');
          // Opcional: Redirigir al login
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

    async getCurrentUser(): Promise<User> {
    const token = localStorage.getItem('token');
    if (!token) {
      // Si no hay token, no hay sesión, rechazamos la promesa inmediatamente.
      return Promise.reject('No token found');
    }
    try {
      // Asume que tienes un endpoint en tu backend que devuelve el usuario basado en el token.
      // Este endpoint es crucial para la persistencia de la sesión.
      const response = await this.api.get<User>('/auth/me');
      return response.data;
    } catch (error) {
      console.error('Failed to get current user with token', error);
      this.logout(); // Limpia el token inválido del storage
      throw new Error('Invalid session');
    }
  }
  
  async login(email: string, password: string): Promise<User> {
    try {
      const response = await this.api.post<{ token: string; user: User }>('/auth/login', {
        email,
        password,
      });
      
      if (response.data.token) {
        localStorage.setItem('token', response.data.token);
      }
      
      return response.data.user;
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw new Error('Error al iniciar sesión');
    }
  }

  async logout() {
    try {
      await this.api.post('/auth/logout');
    } catch (error) {
      console.error('Error en logout:', error);
    }
  }

  async register(email: string, password: string, name?: string) {
    try {
      const response = await this.api.post<{ token: string; user: User }>('/auth/register', {
        email,
        password,
        name,
      });
      
      if (response.data.token) {
        localStorage.setItem('token', response.data.token);
      }

      return response.data.user;
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw new Error('Error al registrar usuario');
    }
  }

  async verifyToken(): Promise<User> {
    try {
      const response = await this.api.get('/auth/verify');
      return response.data.user;
    } catch (error) {
      throw new Error('Token inválido');
    }
  }


  async refreshToken() {
    try {
      const response = await this.api.post('/auth/refresh');
      const newToken = response.data.token;
      
      localStorage.setItem('token', newToken);
      return newToken;
    } catch (error) {
      throw new Error('Error al refrescar token');
    }
  }

  async forgotPassword(email: string) {
    try {
      const response = await this.api.post('/auth/forgot-password', { email });
      return response.data;
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw new Error('Error al enviar email de recuperación');
    }
  }

  async resetPassword(token: string, newPassword: string) {
    try {
      const response = await this.api.post('/auth/reset-password', {
        token,
        password: newPassword,
      });
      return response.data;
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw new Error('Error al restablecer contraseña');
    }
  }
}

export const authService = new AuthService();