import { useState } from "react";

const initialState = {
  email: "",
  password: ""
};

export function LoginForm({ error, loading, onSubmit }) {
  const [form, setForm] = useState(initialState);

  function handleChange(event) {
    const { name, value } = event.target;
    setForm((current) => ({
      ...current,
      [name]: value
    }));
  }

  function handleSubmit(event) {
    event.preventDefault();
    onSubmit(form);
  }

  return (
    <form className="login-form card" onSubmit={handleSubmit}>
      <div className="login-form-head">
        <p className="page-eyebrow">MRAP</p>
        <h1 className="hero-title">Sign in</h1>
        <p className="hero-text">
          Доступ к аналитическому workspace. Используйте backend-учетную запись, чтобы открыть
          dashboard, активы, события и прогнозы.
        </p>
      </div>

      <label className="form-field">
        <span>Email</span>
        <input
          autoComplete="email"
          name="email"
          onChange={handleChange}
          placeholder="user@example.com"
          type="email"
          value={form.email}
        />
      </label>

      <label className="form-field">
        <span>Пароль</span>
        <input
          autoComplete="current-password"
          name="password"
          onChange={handleChange}
          placeholder="Введите пароль"
          type="password"
          value={form.password}
        />
      </label>

      {error ? <div className="error-box">{error}</div> : null}

      <button className="primary-button" disabled={loading} type="submit">
        {loading ? "Выполняем вход..." : "Login"}
      </button>

      <div className="helper-card">
        <strong>Protected analytics workspace</strong>
        <p className="muted-text">
          В MVP клиент хранит сессию локально, а бизнес-данные продолжают запрашиваться из уже
          существующих backend endpoint&apos;ов.
        </p>
      </div>
    </form>
  );
}
