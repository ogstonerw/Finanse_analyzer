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
        <div>
          <p className="page-eyebrow">Market Reaction Analytics Platform</p>
          <h2 className="hero-title">AI platform for market reaction analysis</h2>
          <p className="hero-text">
            Open-source news, macro and market data combined into short-term forecasts и
            интерпретируемый режим рынка.
          </p>
        </div>

        <div className="login-highlight-grid">
          <article className="login-highlight">
            <strong>News-driven signals</strong>
            <p className="muted-text">Лента событий и новостей, привязанная к активам и сигналам.</p>
          </article>
          <article className="login-highlight">
            <strong>Market regime indicator</strong>
            <p className="muted-text">Кризисометр и краткое объяснение текущего состояния рынка.</p>
          </article>
          <article className="login-highlight">
            <strong>Forecast history</strong>
            <p className="muted-text">Последние сгенерированные сигналы с уверенностью и контекстом.</p>
          </article>
        </div>
      </div>

      <LoginForm error={error} loading={loading} onSubmit={handleSubmit} />
    </div>
  );
}
