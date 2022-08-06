import FileBrowser, {FileTreeI} from "./component/FileBrowser";
import {useEffect, useState} from "react";
import {GetFile, GetFileTree, SaveFile} from "./api/file";
import FileEditor, {FileI} from "./component/FileEditor";
import {Header, MenuI} from "./component/Header";
import {MenuVertical} from "./component/MenuVertical";
import Modal from "./component/Modal";
import Input from "./component/Input";

function App() {
    let [fileTree, setFileTree] = useState<FileTreeI>()
    let [currFile, setCurrFile] = useState<FileI>()
    let [newFileModel, setNewFileModel] = useState(false)
    let [drawer, setDrawer] = useState(false)
    let [newFileName, setNewFileName] = useState('')

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

    const onNewFileClick = async () => {
        setNewFileModel(true)
        await SaveFile({path: "index.js", bucket: "", body: ""})
    }

    const switchDrawer = () => {
        setDrawer(!drawer)
    }

    const doNewFile = async () => {
        await SaveFile({
            path: newFileName,
            bucket: "",
            body: "",
        })
        await reloadFileTree()
        setNewFileName('')
        setNewFileModel(false)
    }

    const headMenus: MenuI[] = [{
        key: "file",
        name: "File"
    }]

    const onMenuClick = (m: MenuI) => {
        switch (m.key) {
            case 'file':
                break
            case 'project':
                switchDrawer()
                break
        }
    }

    return (
        <div className="App" data-theme="dark">
            <div className="App-header flex-col h-screen space-y-2 bg-gray-1A1E2A">
                <Header menus={headMenus} onMenuClick={onMenuClick} currFile={currFile}></Header>
                <section className="flex-1 flex h-0 space-x-2">
                    <div className="w-6 ">
                        <MenuVertical menus={[{key: "project", name: "Project"}]}
                                      onMenuClick={onMenuClick}></MenuVertical>
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
                                    onNewFileClick={onNewFileClick}
                                ></FileBrowser>
                            </div>
                        </div>
                    </div>
                </section>
            </div>

            <Modal
                value={newFileModel}
                confirmBtn={"OK"}
                title={"New File"}
                onClose={() => {
                    setNewFileModel(false);
                    setNewFileName('');
                }}
                onConfirm={doNewFile}
                keyEnter={true}
            >
                <Input type="text" value={newFileName} onChange={(e) => {
                    setNewFileName(e.currentTarget.value)
                }}/>
            </Modal>
        </div>
    );
}

export default App;
