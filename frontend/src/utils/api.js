const SIMULATED_USER_KEY = 'simulated_user_id';

export const getSimulatedUserID = () => {
    return localStorage.getItem(SIMULATED_USER_KEY);
};

export const setSimulatedUserID = (id) => {
    if (id) {
        localStorage.setItem(SIMULATED_USER_KEY, id);
    } else {
        localStorage.removeItem(SIMULATED_USER_KEY);
    }
};

export const apiFetch = async (url, options = {}) => {
    const headers = options.headers || {};
    const userID = getSimulatedUserID();

    if (userID) {
        headers['X-Simulated-User-ID'] = userID;
    }

    const config = {
        ...options,
        headers: {
            ...headers,
            'Content-Type': 'application/json',
        },
    };

    const response = await fetch(url, config);

    // Optional: Global error handling (e.g., 401/403 alerts)
    if (response.status === 401) {
        console.warn("Unauthorized access - check simulated user ID");
    }

    return response;
};
