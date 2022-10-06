import axios, {AxiosError} from "axios";
import {serviceAddress} from "../const/const";
import {Repo} from "../particle/PullModal";
import {message} from "../util/Toast";

axios.defaults.baseURL = serviceAddress;
axios.defaults.withCredentials = true

axios.interceptors.response.use(response => {
    return response
}, function (error: AxiosError<any>) {
    message("error", error.response?.data?.msg || error.message)
    return Promise.reject(error);
})

interface LoginParams {
    secret: string
}

export const Login = (params: LoginParams) => axios.post<void>('/api/auth', params);

export const Pull = (repo: Repo) => axios.post<string>('/api/pull');
export const Push = (repo: Repo) => axios.post<string>('/api/push');

export interface Config {
    source: Repo
    deploy: Repo
}

export const GetConfig = () => axios.get<Config>("/api/config")