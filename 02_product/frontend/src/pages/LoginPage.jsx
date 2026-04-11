import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { LoginForm } from "../components/LoginForm";

export function LoginPage({ onLogin }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const navigate = useNavigate();

  async function handleSubmit(credentials) {
    setLoading(true);
    setError("");

    try {
      const response = await api.login(credentials);
      onLogin({
        email: response.email,
        expires_at: response.expires_at,
        session_token: response.session_token,
        user_id: response.user_id
      });
      navigate("/dashboard", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось выполнить вход");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-page">
      <div className="login-visual">
        <p className="eyebrow">React MVP</p>
        <h2 className="hero-title">Интерфейс для демонстрации платформы анализа рынка</h2>
        <ul className="feature-list">
          <li>dashboard со сводкой рынка и кризисометром;</li>
          <li>каталог активов и карточка инструмента;</li>
          <li>новости, события и последние прогнозы;</li>
          <li>подключение к уже существующему backend API.</li>
        </ul>
      </div>

      <LoginForm error={error} loading={loading} onSubmit={handleSubmit} />
    </div>
  );
}
