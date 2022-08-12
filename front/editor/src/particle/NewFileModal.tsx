import ReactDOM from "react-dom/client";
import {ReactNode, useCallback, useState} from "react";
import {useDropzone} from "react-dropzone";
import Input from "../component/Input";
import Modal from "../component/Modal";

interface ConfirmParams {
    newFileInfo?: NewFileInfo
    onClose: ()=>void
    onConfirm: (filename: string, uploadFiles: File[])=> Promise<void>
}

export interface NewFileInfo {
    isDir: boolean,
    parentPath: string,
}

// 弹出框
export default function NewFileModal(p: ConfirmParams) {
    const [newFileName, setNewFileName] = useState('')
    const [uploadFiles, setUploadFiles] = useState<File[]>([])

    const onDrop = useCallback((acceptedFiles: File[]) => {
        setUploadFiles(acceptedFiles)
    }, [])
    const {getRootProps, getInputProps, isDragActive} = useDropzone({onDrop, noClick: true, noKeyboard: true})

    const previewFile = (fs: File[]): string => {
        // @ts-ignore
        return fs.map(i => (i.path)).join("\n")
    }

    const onCloseNewFile = () => {
        p.onClose()
        setNewFileName('');
        setUploadFiles([])
    }
    const doNewFile = async () => {
        await p.onConfirm(newFileName, uploadFiles)
        setNewFileName('')
        setUploadFiles([])
    }
    const m2 = <Modal
        value={!!p.newFileInfo}
        confirmBtn={"OK"}
        title={p.newFileInfo?.isDir ? "New Directory" : "New File"}
        onClose={onCloseNewFile}
        onConfirm={doNewFile}
        keyEnter={true}
    >
        {/*{...getRootProps()}*/}
        <div {...getRootProps()}>
            {
                isDragActive ? (
                        <p className="btn btn-outline btn-sm btn-block">Upload files or drag some files
                            here</p>
                    ) :
                    (
                        uploadFiles.length ? (
                            <div className="form-control">
                                <label className="label">
                                        <span
                                            className="label-text">{p.newFileInfo?.parentPath ? `Create in '${p.newFileInfo?.parentPath}' directory` : ''}</span>
                                </label>
                                <textarea
                                    disabled
                                    className="textarea textarea-bordered w-full" placeholder="Bio"
                                    value={previewFile(uploadFiles)}
                                />
                            </div>
                        ) : (
                            <Input
                                label={p.newFileInfo?.parentPath ? `Create in '${p.newFileInfo?.parentPath}' directory` : ''}
                                className="mt-3"
                                autoFocus={true}
                                type="text"
                                value={newFileName}
                                onChange={(e) => {
                                    setNewFileName(e.currentTarget.value)
                                }}/>
                        )

                    )

            }
        </div>

    </Modal>

    return m2
}