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
    body?: string
}

interface UploadFilesParams {
    path: string
    files: File[]
}

interface DeleteFileParams {
    path: string
    is_dir: boolean
}

export const SaveFile = (params: SaveFileParams) =>
    axios.put<void>('/api/file', params);

export const DeleteFile = (params: DeleteFileParams) =>
    axios.delete<void>('/api/file', {
        params
    });

export const CreateDirectory = (params: SaveFileParams) =>
    axios.put<void>('/api/directory', params);

export const UploadFiles = (params: UploadFilesParams) => {
    const forms = new FormData()
    const configs = {
        headers: {'Content-Type': 'multipart/form-data'}
    };
    params.files.forEach(i => {
        // @ts-ignore
        console.log('i.path', i.path)
        // @ts-ignore
        forms.append('file', i, i.path)
    })
    forms.append("path", params.path)

    return axios.put<void>('/api/file/upload', forms, configs);
}