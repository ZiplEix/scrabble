import axios from "axios";
import { user } from "./stores/user";
import { env } from "$env/dynamic/public";
import { goto } from "$app/navigation";

export const api = axios.create({
    baseURL: env.PUBLIC_API_BASE_URL,
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

api.interceptors.response.use(
    response => response,
    error => {
        if (error.response && error.response.status === 401) {
            user.set(null);
            goto('/');
        }
        return Promise.reject(error);
    }
)
