import FileBrowser, {FileTreeI} from "./component/FileBrowser";
import {useEffect, useState} from "react";
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
import FileEditor, {FileI} from "./component/FileEditor";
import {Header, MenuI} from "./component/Header";
import {MenuVertical} from "./component/MenuVertical";
import Confirm from "./component/Confirm";
import NewFileModal, {NewFileInfo} from "./particle/NewFileModal";
import PublishModal from "./particle/PublishModal";
import Ws from "./util/ws";

function App() {
    const [pid, setPid] = useState(1)
    const [workspace, setWorkspace] = useState<'project' | 'theme'>('project')
    const [fileTreeProject, setFileTreeProject] = useState<FileTreeI>()
    const [fileTreeTheme, setFileTreeTheme] = useState<FileTreeI>()
    const [currFile, setCurrFile] = useState<FileI>()
    const [newFileInfo, setNewFileInfo] = useState<NewFileInfo>()
    const [ws, setWs] = useState<any>(null)
    const [showPublishModal, setShowPublishModal] = useState(false)
    const [drawer, setDrawer] = useState(false)
    const bucket = workspace

    const reloadFileTree = async () => {
        const ft = await GetFileTree({project_id: pid, path: "", bucket: 'project'})
        setFileTreeProject(ft.data)
        {
            const ft = await GetFileTree({project_id: pid, path: "", bucket: 'theme'})
            setFileTreeTheme(ft.data)
        }
    }
    useEffect(() => {
        (reloadFileTree)();
    }, [])

    const onFileChange = async (body: string) => {
        await SaveFile({project_id: pid, path: currFile?.path!, bucket: bucket, body: body})
    }

    const onFileClick = async (f: FileI) => {
        setCurrFile(f)

        if (!f.is_dir) {
            const nf = await GetFile({project_id: pid, path: f.path, bucket: bucket})
            setCurrFile(nf.data)
        }
    }


    useEffect(() => {
        setWs(new Ws())
    }, [])

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
            case 'project':
                setWorkspace("project")
                switchDrawer()
                break
            case 'theme':
                setWorkspace("theme")
                break
            case 'publish':
                setShowPublishModal(true)
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
                break
        }
    }

    // @ts-ignore
    return (
        <div className="App" data-theme="dark">
            <div className="App-header flex-col h-screen space-y-2 bg-gray-1A1E2A">
                <Header menus={headMenus} onMenuClick={onTopMenu} currFile={currFile}></Header>
                <section className="flex-1 flex h-0 ">
                    <div className="w-6 ">
                        <MenuVertical
                            menus={[
                                {key: "project", name: "Project"},
                                {key: "theme", name: "Theme"},
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
                            <div className="bg-gray-272C38 rounded-lg h-full  p-2">
                                <FileEditor file={currFile} onChange={onFileChange}/>
                            </div>
                        </div>
                        <div className="drawer-side" style={{"height": '100%', 'overflowY': "auto"}}>
                            <label onClick={() => setDrawer(false)} className="drawer-overlay "></label>
                            <div className="menu w-60 flex flex-col mr-2 bg-gray-272C38 rounded-lg overflow-y-auto">
                                <>
                                    <div style={{display: workspace === 'project' ? '' : 'none'}}>
                                        <FileBrowser
                                            tree={fileTreeProject}
                                            currFile={currFile}
                                            onFileClick={onFileClick}
                                            onMenu={onFileMenu}
                                        ></FileBrowser>
                                    </div>

                                    <div style={{display: workspace === 'theme' ? '' : 'none'}}>
                                        <FileBrowser
                                            tree={fileTreeTheme}
                                            currFile={currFile}
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
                ws={ws}
            ></PublishModal>
        </div>
    );
}

export default App;
