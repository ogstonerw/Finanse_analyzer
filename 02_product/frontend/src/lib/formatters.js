export function formatDateTime(value) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "—";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date);
}

export function formatTime(value) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}

export function formatNumber(value, digits = 2) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return "-";
  }

  return new Intl.NumberFormat("ru-RU", {
    maximumFractionDigits: digits,
    minimumFractionDigits: 0
  }).format(Number(value));
}

export function formatPercent(value, options = {}) {
  const { digits = 2, multiplier = 1 } = options;
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return "—";
  }

  return `${formatNumber(Number(value) * multiplier, digits)}%`;
}

export function formatPrice(value, currency = "") {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return "-";
  }

  const normalizedCurrency = String(currency || "").toUpperCase();
  const digits = Math.abs(Number(value)) < 1 ? 3 : 2;
  const formatted = formatNumber(value, digits);

  switch (normalizedCurrency) {
    case "RUB":
      return `${formatted} ₽`;
    case "USD":
      return `$${formatted}`;
    case "EUR":
      return `€${formatted}`;
    default:
      return normalizedCurrency ? `${formatted} ${normalizedCurrency}` : formatted;
  }
}

export function getDirectionLabel(direction) {
  switch (direction) {
    case "positive":
      return "Bullish";
    case "negative":
      return "Bearish";
    case "neutral":
      return "Neutral";
    default:
      return direction || "-";
  }
}

export function getTrendLabel(direction) {
  switch (direction) {
    case "up":
      return "Up";
    case "down":
      return "Down";
    case "flat":
      return "Flat";
    default:
      return "-";
  }
}

export function getRegimeLabel(label) {
  switch (label) {
    case "stable":
      return "Стабильный";
    case "moderate_tension":
      return "Умеренное напряжение";
    case "elevated_stress":
      return "Повышенный стресс";
    case "pre_crisis":
      return "Предкризисный";
    case "crisis":
      return "Кризис";
    default:
      return label || "-";
  }
}

export function getStrengthLabel(value) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return "-";
  }

  const normalized = Number(value);

  if (normalized >= 0.67) {
    return "Strong";
  }

  if (normalized >= 0.34) {
    return "Medium";
  }

  return "Low";
}

export function getConfidenceLabel(value) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return "-";
  }

  const normalized = Number(value);

  if (normalized >= 0.7) {
    return "High";
  }

  if (normalized >= 0.4) {
    return "Medium";
  }

  return "Low";
}

export function getAssetTypeLabel(value) {
  const normalized = String(value || "").toLowerCase();

  switch (normalized) {
    case "stock":
    case "stocks":
    case "share":
    case "shares":
      return "Stocks";
    case "index":
    case "indices":
      return "Index";
    case "commodity":
    case "commodities":
      return "Commodities";
    default:
      return value || "-";
  }
}

export function getEventTypeLabel(value) {
  if (!value) {
    return "Event";
  }

  return String(value)
    .replace(/_/g, " ")
    .replace(/\b\w/g, (match) => match.toUpperCase());
}

export function getDirectionClassName(direction) {
  switch (direction) {
    case "positive":
      return "positive";
    case "negative":
      return "negative";
    case "neutral":
      return "neutral";
    default:
      return "muted";
  }
}
