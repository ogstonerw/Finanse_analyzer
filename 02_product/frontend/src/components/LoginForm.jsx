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
      <div>
        <p className="eyebrow">Market AI Platform</p>
        <h1 className="hero-title">Вход в MVP интерфейс</h1>
        <p className="hero-text">
          Используйте учетную запись backend, чтобы открыть dashboard, кризисометр, активы и
          прогнозы.
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
        {loading ? "Входим..." : "Войти"}
      </button>

      <p className="muted-text">
        MVP не реализует полноценную защищенную клиентскую авторизацию: после успешного login
        токен просто сохраняется локально для демонстрации навигации.
      </p>
    </form>
  );
}
