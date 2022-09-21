import FileBrowser, {FileTreeI} from "./component/FileBrowser";
import {useCallback, useEffect, useMemo, useState} from "react";
import {
    CreateDirectory,
    CreateFile,
    DeleteFile,
    GetFile,
    GetFileTree,
    Publish, Pull, Push,
    SaveFile,
    UploadFiles
} from "./api/file";
import FileEditor, {FileI} from "./component/FileEditor";
import {Header, MenuI} from "./component/Header";
import {MenuVertical} from "./component/MenuVertical";
import Confirm from "./component/Confirm";
import NewFileModal, {NewFileInfo} from "./particle/NewFileModal";
import PublishModal from "./particle/PublishModal";
import Ws from "./util/ws";
import {ShowPopupMenu} from "./util/popupMenu";
import {DownloadIcon} from "./icon";
import debounce from "lodash/debounce";
import ProcessModal from "./particle/ProcessModal";
import LoginModal from "./particle/LoginModal";
import {Login} from "./api/base";
import {AxiosError} from "axios";
import {set} from "lodash";

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

    useEffect(() => {
        const convertStyle = () => {
            const height = window.innerHeight;
            // alert(height)
            // document.body.style.height = `${height}px`;
        }
        window.addEventListener("resize", convertStyle);
        setInterval(() => {
            convertStyle()
        }, 1000)
        convertStyle()
    })

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

    const onTopMenu = (m: MenuI) => {
        switch (m.key) {
            case 'publish':
                setShowPublishModal(true)
                return
            case 'menu':
                ShowPopupMenu({
                    x: 20,
                    y: 20,
                    menu: [
                        {
                            name: <div className={"flex space-x-2"}>
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                     strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                    <path strokeLinecap="round" strokeLinejoin="round"
                                          d="M7.5 7.5h-.75A2.25 2.25 0 004.5 9.75v7.5a2.25 2.25 0 002.25 2.25h7.5a2.25 2.25 0 002.25-2.25v-7.5a2.25 2.25 0 00-2.25-2.25h-.75m-6 3.75l3 3m0 0l3-3m-3 3V1.5m6 9h.75a2.25 2.25 0 012.25 2.25v7.5a2.25 2.25 0 01-2.25 2.25h-7.5a2.25 2.25 0 01-2.25-2.25v-.75"/>
                                </svg>
                                <span>Update Project</span>
                            </div>,
                            key: "update project"
                        },
                        {
                            name: <div className={"flex space-x-2"}>
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                     strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                    <path strokeLinecap="round" strokeLinejoin="round"
                                          d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5"/>
                                </svg>
                                <span>Push</span>
                            </div>,
                            key: "push"
                        },
                    ],
                    onClick: async (m) => {
                        switch (m.key) {
                            case "update project":
                                let r = await Pull()

                                setProcessModal({
                                    title: "update",
                                    wsKey: r.data,
                                })
                                break
                            case "push":
                                let rr = await Push()

                                setProcessModal({
                                    title: "update",
                                    wsKey: rr.data,
                                })
                        }
                    },
                    id: "",
                    mask: true,
                })
        }
        onLeftTab(m)
        return
    }

    let login = async (secret: string) => {
        await Login({secret})
        window.location = window.location
        return
    }

    useEffect(() => {
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

    // @ts-ignore
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
            <ProcessModal
                onClose={() => {
                    setProcessModal(undefined)
                }}
                show={!!processModal}
                onConfirm={doPublish}
                wsKey={processModal?.wsKey}
            ></ProcessModal>
            <LoginModal onConfirm={login} show={loginModal}></LoginModal>
        </div>
    );
}

export default App;
