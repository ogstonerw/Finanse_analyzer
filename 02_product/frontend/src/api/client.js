const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || "").replace(/\/$/, "");

function buildUrl(path) {
  return API_BASE_URL ? `${API_BASE_URL}${path}` : path;
}

async function readPayload(response) {
  const contentType = response.headers.get("Content-Type") || "";
  if (!contentType.includes("application/json")) {
    return null;
  }

  return response.json();
}

export async function apiRequest(path, options = {}) {
  const { method = "GET", body, headers = {}, signal } = options;
  const requestHeaders = { ...headers };
  const requestOptions = { method, headers: requestHeaders, signal };

  if (body !== undefined) {
    requestHeaders["Content-Type"] = "application/json";
    requestOptions.body = JSON.stringify(body);
  }

  let response;

  try {
    response = await fetch(buildUrl(path), requestOptions);
  } catch (error) {
    throw new Error(
      "Не удалось подключиться к backend API. Проверьте, что backend запущен и доступен по локальному адресу."
    );
  }

  const payload = await readPayload(response);

  if (!response.ok) {
    const message =
      payload && typeof payload === "object" && "error" in payload
        ? payload.error
        : `Request failed with status ${response.status}`;
    throw new Error(message);
  }

  return payload;
}

export const api = {
  login(credentials) {
    return apiRequest("/api/v1/auth/login", { body: credentials, method: "POST" });
  },
  getDashboardSummary() {
    return apiRequest("/api/v1/dashboard/summary");
  },
  getCurrentRegime() {
    return apiRequest("/api/v1/regime/current");
  },
  getAssets() {
    return apiRequest("/api/v1/assets");
  },
  getAsset(ticker) {
    return apiRequest(`/api/v1/assets/${ticker}`);
  },
  getAssetPrices(ticker) {
    return apiRequest(`/api/v1/assets/${ticker}/prices`);
  },
  getAssetIndicators(ticker) {
    return apiRequest(`/api/v1/assets/${ticker}/indicators`);
  },
  getNews() {
    return apiRequest("/api/v1/news");
  },
  getEvents() {
    return apiRequest("/api/v1/events");
  },
  getLatestForecast() {
    return apiRequest("/api/v1/forecasts/latest");
  }
};
