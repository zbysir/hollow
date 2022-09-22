import axios from "axios";
import {serviceAddress} from "../const/const";
import {Repo} from "../particle/PullModal";
import exp from "constants";

axios.defaults.baseURL = serviceAddress;
axios.defaults.withCredentials = true

// axios.interceptors.response.use(response => {
//     let resData = response.data
//     if (resData.code === -1) {
//         return Promise.reject(new Error(resData.msg))
//     } else {
//         return resData.data
//     }
// })

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