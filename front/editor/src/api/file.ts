import axios from 'axios'
import {FileTreeI} from "../component/FileBrowser";
import {FileI} from "../component/FileEditor";
import {serviceAddress} from "../const/const";
import {Repo} from "../particle/PullModal";


interface GetFileTreeParams {
    path: string
    bucket: string
    project_id: number
}

export const GetFileTree = (params: GetFileTreeParams) => axios.get<FileTreeI>('/api/file/tree', {
    params: params
});

export const GetFile = (params: GetFileTreeParams) => axios.get<FileI>('/api/file', {
    params: params
});


interface PublishParams {
    project_id: number
}
interface PublishRsp {
    key: string
}
export const Publish = (params: PublishParams) => axios.post<string>('/api/publish', params);

interface SaveFileParams {
    project_id: number
    path: string
    bucket: string
    body?: string
}

interface UploadFilesParams {
    project_id: number
    path: string
    files: File[]
    bucket: string
}

interface DeleteFileParams {
    project_id: number
    path: string
    is_dir: boolean
    bucket: string
}

// 修改文件
export const SaveFile = (params: SaveFileParams) =>
    axios.put<void>('/api/file', params);

// 创建文件
export const CreateFile = (params: SaveFileParams) =>
    axios.post<void>('/api/file', params);

export const DeleteFile = (params: DeleteFileParams) =>
    axios.delete<void>('/api/file', {
        params
    });

export const CreateDirectory = (params: SaveFileParams) =>
    axios.post<void>('/api/directory', params);

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
    forms.append("bucket", params.bucket)
    forms.append("project_id", params.project_id + '')

    return axios.put<string[]>('/api/file/upload', forms, configs);
}
