const API_URL = "http://localhost:8080";

export const api = {
    signup: async (username: string, email: string, password: string) => {
        const response = await fetch(`${API_URL}/sign-up`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, email, password }),
        });
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || "Signup failed");
        }
        return response.json();
    },

    login: async (email: string, password: string) => {
        const response = await fetch(`${API_URL}/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email, password }),
        });
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || "Login failed");
        }
        return response.json();
    },

    getRooms: async (token: string) => {
        const response = await fetch(`${API_URL}/api/rooms?token=${token}`, {
            headers: { "Authorization": `Bearer ${token}` }
        });
        if (!response.ok) throw new Error("Failed to fetch rooms");
        return response.json();
    },

    createRoom: async (name: string, token: string) => {
        const response = await fetch(`${API_URL}/api/room/create?token=${token}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${token}`
            },
            body: JSON.stringify({ name }),
        });
        if (!response.ok) throw new Error("Failed to create room");
        return response.json();
    },

    joinRoom: async (roomId: number, token: string) => {
        const response = await fetch(`${API_URL}/api/room/join?token=${token}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${token}`
            },
            body: JSON.stringify({ room_id: roomId }),
        });
        if (!response.ok) throw new Error("Failed to join room");
        return response.json();
    }
};
