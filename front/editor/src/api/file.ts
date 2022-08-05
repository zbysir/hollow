import axios from 'axios'
import {FileTreeI} from "../component/FileBrowser";
import {FileI} from "../component/FileEditor";

axios.defaults.baseURL = '//localhost:9090';

interface GetFileTreeParams {
    path: string
    bucket: string
}

export const GetFileTree = (params: GetFileTreeParams) => axios.get<FileTreeI>('/api/file/tree', {
    params: params
});

export const GetFile = (params: GetFileTreeParams) => axios.get<FileI>('/api/file', {
    params: params
});

interface SaveFileParams {
    path: string
    bucket: string
    body: string
}

export const SaveFile = (params: SaveFileParams) =>
    axios.put<void>('/api/file', params);