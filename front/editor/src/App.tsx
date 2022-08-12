import FileBrowser, {FileTreeI} from "./component/FileBrowser";
import {useCallback, useEffect, useMemo, useState} from "react";
import {CreateDirectory, DeleteFile, GetFile, GetFileTree, SaveFile, UploadFiles} from "./api/file";
import FileEditor, {FileI} from "./component/FileEditor";
import {Header, MenuI} from "./component/Header";
import {MenuVertical} from "./component/MenuVertical";
import Confirm from "./component/Confirm";
import NewFileModal, {NewFileInfo} from "./particle/NewFileModal";


function App() {
    const [fileTree, setFileTree] = useState<FileTreeI>()
    const [currFile, setCurrFile] = useState<FileI>()
    const [newFileInfo, setNewFileInfo] = useState<NewFileInfo>()
    const [drawer, setDrawer] = useState(false)

    const reloadFileTree = async () => {
        const ft = await GetFileTree({path: "", bucket: ""})
        // ft.data.items!.push(...ft.data.items!)
        // ft.data.items!.push(...ft.data.items!)
        setFileTree(ft.data)
    }
    useEffect(() => {
        (reloadFileTree)();
    }, [])
    let mock = fileTree

    const onFileChange = async (body: string) => {
        await SaveFile({path: currFile?.path!, bucket: "", body: body})
    }

    const onFileClick = async (f: FileI) => {
        setCurrFile(f)

        if (!f.is_dir) {
            const nf = await GetFile({path: f.path, bucket: ""})
            setCurrFile(nf.data)
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
                    await DeleteFile({path: f.path, is_dir: f.is_dir})
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
                files: uploadFiles,
                path: newFileInfo?.parentPath!,
            })
        } else {
            const path = newFileInfo?.parentPath + "/" + newFileName
            if (newFileInfo?.isDir) {
                await CreateDirectory({
                    path: path,
                    bucket: "",
                    body: "",
                })
            } else {
                await SaveFile({
                    path: path,
                    bucket: "",
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

    const onLeftMenu = (m: MenuI) => {
        switch (m.key) {
            case 'file':
                break
            case 'project':
                switchDrawer()
                break
        }
    }

    // @ts-ignore
    return (
        <div className="App" data-theme="dark">
            <div className="App-header flex-col h-screen space-y-2 bg-gray-1A1E2A">
                <Header menus={headMenus} onMenuClick={onLeftMenu} currFile={currFile}></Header>
                <section className="flex-1 flex h-0 space-x-2">
                    <div className="w-6 ">
                        <MenuVertical menus={[
                            {key: "project", name: "Project"},
                            {key: "theme", name: "Theme"},
                        ]}
                                      onMenuClick={onLeftMenu}></MenuVertical>
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
                                <FileBrowser
                                    tree={mock}
                                    currFile={currFile}
                                    onFileClick={onFileClick}
                                    onMenu={onFileMenu}
                                ></FileBrowser>
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
        </div>
    );
}

export default App;
