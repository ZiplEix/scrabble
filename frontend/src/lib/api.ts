import axios from "axios";
import { user } from "./stores/user";

export const api = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8888',
    withCredentials: false,
    headers: {
        'Content-Type': 'application/json'
    }
})

user.subscribe(($user) => {
	if ($user?.token) {
		api.defaults.headers.common['Authorization'] = `Bearer ${$user.token}`;
	} else {
		delete api.defaults.headers.common['Authorization'];
	}
});
