export function formatDateTime(value) {
  if (!value) {
    return "—";
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

export function formatNumber(value, digits = 2) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return "—";
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

export function getDirectionLabel(direction) {
  switch (direction) {
    case "positive":
      return "Позитивный";
    case "negative":
      return "Негативный";
    case "neutral":
      return "Нейтральный";
    default:
      return direction || "—";
  }
}

export function getTrendLabel(direction) {
  switch (direction) {
    case "up":
      return "Рост";
    case "down":
      return "Снижение";
    case "flat":
      return "Боковик";
    default:
      return "—";
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
      return label || "—";
  }
}
