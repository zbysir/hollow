import axios from "axios";
import {serviceAddress} from "../const/const";

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