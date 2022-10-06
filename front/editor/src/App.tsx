import FileBrowser, {FileTreeI} from "./component/FileBrowser";
import {useCallback, useEffect, useMemo, useState} from "react";
import {
    CreateDirectory,
    CreateFile,
    DeleteFile,
    GetFile,
    GetFileTree,
    Publish,
    SaveFile,
    UploadFiles
} from "./api/file";
import {GetConfig, Login, Pull, Push} from "./api/base";

import FileEditor, {FileI} from "./component/FileEditor";
import {Header, MenuI} from "./component/Header";
import {MenuVertical} from "./component/MenuVertical";
import Confirm from "./component/Confirm";
import NewFileModal, {NewFileInfo} from "./particle/NewFileModal";
import PublishModal from "./particle/PublishModal";
import {ShowPopupMenu} from "./util/popupMenu";
import debounce from "lodash/debounce";
import ProcessModal from "./particle/ProcessModal";
import LoginModal from "./particle/LoginModal";
import {AxiosError} from "axios";
import PullModal, {Repo} from "./particle/PullModal";
import ThemeModal from "./particle/ThemeModal";
import {Toast} from "./util/Toast";

// FileStatus 可以被序列化，刷新页面恢复
export interface FileStatus {
    modifiedFiles: FileI[]
    currFile?: FileTreeI
    openedDir?: FileI[]
}

interface ProcessModalI {
    title: string
    wsKey?: string
}

function UseStorage<T>(key: string, initVal: T): [T, (t: T) => void] {
    const raw = localStorage.getItem(key)

    const [value, setValue] = useState<T>(raw ? JSON.parse(raw) : initVal)
    const updater = useCallback(
        (updatedValue: T) => {
            localStorage.setItem(key, JSON.stringify(updatedValue))
            setValue(updatedValue);
        },
        [key],
    );

    return [value, updater]
}

interface PullModalData {
    repo?: Repo
}

function App() {
    const [pid, setPid] = useState(1)
    const [workspace, setWorkspace] = useState<'project' | 'theme'>('project')
    const [fileTreeProject, setFileTreeProject] = useState<FileTreeI>()
    const [newFileInfo, setNewFileInfo] = useState<NewFileInfo>()
    const [showPublishModal, setShowPublishModal] = useState(false)
    const [drawer, setDrawer] = useState(false)
    const [fileStatus, setFileStatus] = UseStorage<FileStatus>("file_status", {modifiedFiles: []})
    const bucket = workspace
    const [processModal, setProcessModal] = useState<ProcessModalI>()
    const [loginModal, setLoginModal] = useState<boolean>(false)

    const [pullModal, setPullModal] = useState<PullModalData>()

    const setFileStatusFileModified = (fileStatus: FileStatus, f: FileI, modify: boolean) => {
        const newStatus = {...fileStatus}
        // console.log('fileStatus.modifiedFiles', fileStatus.modifiedFiles)
        const idx = fileStatus.modifiedFiles.findIndex(i => i.path === f.path)
        if (idx === -1) {
            if (modify) {
                newStatus.modifiedFiles.push(f)
            } else {
                return
            }
        } else {
            if (!modify) {
                newStatus.modifiedFiles.splice(idx, 1)
            } else {
                return
            }
        }

        setFileStatus(newStatus)
    }

    const reloadFileTree = async () => {
        const ft = await GetFileTree({project_id: pid, path: "", bucket: 'project'})
        setFileTreeProject(ft.data)
        {
            // const ft = await GetFileTree({project_id: pid, path: "", bucket: 'theme'})
            // setFileTreeTheme(ft.data)
        }
    }
    useEffect(() => {
        (reloadFileTree)();
    }, [])

    const onFileChange = useCallback((f: FileI) => {
        console.log('onFileChange', f.path)
        setFileStatusFileModified(fileStatus, f!, true)
        // 自动保存
        debounceSave(fileStatus, f)
    }, [fileStatus])

    const onFileSave = async (fileStatus: FileStatus, f: FileI) => {
        console.log('onFileSave', f.path)
        setFileStatusFileModified(fileStatus, f, false)
        await SaveFile({project_id: pid, path: f?.path!, bucket: bucket, body: f.body})
    }


    const debounceSave = useMemo(() => {
        return debounce(async (fileStatus: FileStatus, f: FileI) => {
            await onFileSave(fileStatus, f)
        }, 1000)
    }, [])

    const onFileClick = async (f: FileI) => {
        if (!f.is_dir) {
            // const nf = await GetFile({project_id: pid, path: f.path, bucket: bucket})
            setFileStatus({
                ...fileStatus,
                currFile: f,
            })
        } else {
            const newStatus = {...fileStatus}
            if (!newStatus.openedDir) {
                newStatus.openedDir = []
            }

            const idx = fileStatus.openedDir?.findIndex(i => i.path === f.path)
            if (idx !== undefined && idx >= 0) {
                newStatus.openedDir.splice(idx, 1)
            } else {
                newStatus.openedDir.push(f)
            }

            setFileStatus(newStatus)
        }
    }

    const onFileMenu = async (m: MenuI, f: FileTreeI) => {
        switch (m.key) {
            case 'new file':
                setNewFileInfo({
                    isDir: false,
                    parentPath: f.dir_path,
                })
                break
            case 'new directory':
                setNewFileInfo({
                    isDir: true,
                    parentPath: f.dir_path,
                })
                break
            case 'delete':
                const r = await Confirm({
                    title: "Delete",
                    children: (f.is_dir ? (<span>delete directory '{f.path}'？</span>) :
                        <span>delete file '{f.path}'？</span>)
                })
                if (r.ok) {
                    await DeleteFile({project_id: pid, path: f.path, is_dir: f.is_dir, bucket: bucket})
                    await reloadFileTree()
                }
        }
    }

    const switchDrawer = () => {
        setDrawer(!drawer)
    }

    const doNewFile = async (newFileName: string, uploadFiles: File[]) => {
        if (uploadFiles.length !== 0) {
            await UploadFiles({
                project_id: pid,
                files: uploadFiles,
                path: newFileInfo?.parentPath!,
                bucket: bucket,
            })
        } else {
            const path = newFileInfo?.parentPath + "/" + newFileName
            if (newFileInfo?.isDir) {
                await CreateDirectory({
                    project_id: pid,
                    path: path,
                    bucket: bucket,
                    body: "",
                })
            } else {
                await CreateFile({
                    project_id: pid,
                    path: path,
                    bucket: bucket,
                    body: "",
                })
            }
        }

        await reloadFileTree()
        setNewFileInfo(undefined)
    }
    const onCloseNewFile = () => {
        setNewFileInfo(undefined);
    }

    const headMenus: MenuI[] = [{
        key: "file",
        name: "File"
    }]

    const onLeftTab = (m: MenuI) => {
        switch (m.key) {
            case 'files':
                setWorkspace("project")
                switchDrawer()
                break
            case 'theme':
                setWorkspace("theme")
                break

        }
    }

    const doPublish = async () => {
        await Publish({
            project_id: pid,
        })
    }

    const onTopMenu = async (m: MenuI) => {
        switch (m.key) {
            case 'publish':
                setShowPublishModal(true)
                return
            case "update project":
                const r = await GetConfig()
                setPullModal({repo: r.data.source})
                break
            case "push":
                const r1 = await GetConfig()
                let rr = await Push(r1.data.source)
                setProcessModal({
                    title: "update",
                    wsKey: rr.data,
                })
        }
        onLeftTab(m)
        return
    }

    const doPull = async (repo: Repo) => {
        let r = await Pull(repo)
        setPullModal(undefined)
        setProcessModal({
            title: "update",
            wsKey: r.data,
        })
    }

    let login = async (secret: string) => {
        await Login({secret})
        window.location = window.location
        return
    }

    useEffect(() => {
        console.log('app');

        (async function () {
            try {
                let r = await Login({secret: ''})
                console.log(r)
            } catch (e) {
                if (e instanceof AxiosError) {
                    console.log(e.response?.data.code, e.response?.data.code == 401)
                    if (e.response?.data.code == 401) {
                        setLoginModal(true)
                    }
                }
            }
        })()
    }, [])

    return (
        <div id="app" className=" h-full" data-theme="dark">
            <div className="flex flex-col space-y-2 bg-gray-1A1E2A h-full">
                <Header menus={headMenus} onMenuClick={onTopMenu} currFile={fileStatus.currFile} drawer={drawer}
                        fileStatus={fileStatus}></Header>
                <section className="flex-1 flex h-0 relative">
                    <div className="absolute z-10 p-1 pl-0 pt-0 " style={{left: 0, top: 0}}>
                        <MenuVertical
                            menus={[
                                {key: "files", name: "Files"},
                            ]}
                            activeKey={workspace}
                            onMenuClick={onLeftTab}></MenuVertical>
                    </div>
                    <div className="drawer drawer-mobile h-auto flex-1">
                        <input
                            type="checkbox"
                            checked={drawer}
                            onChange={() => {
                            }}
                            className="drawer-toggle"/>
                        <div className="drawer-content h-full">
                            <div className=" rounded-lg h-full overflow-hidden">
                                <FileEditor file={fileStatus.currFile} onChange={onFileChange} onSave={async (f) => {
                                    await onFileSave(fileStatus, f)
                                }}/>
                            </div>
                        </div>
                        <div className="drawer-side" style={{"height": '100%', 'overflowY': "auto"}}>
                            <label onClick={() => setDrawer(false)} className="drawer-overlay "></label>
                            <div
                                className="menu w-60 flex flex-col mr-2 bg-gray-272C38 rounded-lg overflow-y-auto overflow-x-auto">
                                <>
                                    <div style={{display: workspace === 'project' ? '' : 'none'}}>
                                        <FileBrowser
                                            tree={fileTreeProject}
                                            status={fileStatus}
                                            onFileClick={onFileClick}
                                            onMenu={onFileMenu}
                                        ></FileBrowser>
                                    </div>
                                </>

                            </div>
                        </div>
                    </div>
                </section>
            </div>

            {/* New file Modal */}
            <NewFileModal
                onClose={onCloseNewFile}
                onConfirm={doNewFile}
                newFileInfo={newFileInfo}
            ></NewFileModal>
            {/* Publish Modal */}
            <PublishModal
                onClose={() => {
                    setShowPublishModal(false)
                }}
                show={showPublishModal}
                onConfirm={doPublish}
            ></PublishModal>

            {pullModal ?
                <PullModal
                    repo={pullModal?.repo}
                    show={!!pullModal}
                    onClose={() => {
                        setPullModal(undefined)
                    }}
                    onConfirm={async (v) => {
                        await doPull(v)
                    }}></PullModal> : null}

            {processModal ?
                <ProcessModal
                    onClose={() => {
                        setProcessModal(undefined)
                    }}
                    show={true}
                    onConfirm={doPublish}
                    wsKey={processModal?.wsKey}
                ></ProcessModal> : null}

            <LoginModal onConfirm={login} show={loginModal}></LoginModal>

            <Toast/>

        </div>
    );
}

export default App;
