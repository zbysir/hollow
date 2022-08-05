import FileBrowser, {FileTreeI} from "./component/FileBrowser";
import {useEffect, useState} from "react";
import {GetFile, GetFileTree, SaveFile} from "./api/file";
import FileEditor, {FileI} from "./component/FileEditor";

function App() {
    let [fileTree, setFileTree] = useState<FileTreeI>()
    let [currFile, setCurrFile] = useState<FileI>()
    let [newFileModel, setNewFileModel] = useState(false)
    let [drawer, setDrawer] = useState(false)


    useEffect(() => {
        (async function anyNameFunction() {
            const ft = await GetFileTree({path: "", bucket: ""})
            ft.data.items!.push(...ft.data.items!)
            ft.data.items!.push(...ft.data.items!)
            setFileTree(ft.data)
        })();
    }, [])
    console.log('fileTree', fileTree)
    let mock = fileTree

    const onFileChange = async (body: string) => {
        await SaveFile({path: currFile?.path!, bucket: "", body: body})
    }

    const onFileClick = async (f: FileI) => {
        console.log('onFileClick', f)
        const nf = await GetFile({path: f.path, bucket: ""})
        setCurrFile(nf.data)
    }

    const onNewFileClick = async () => {
        setNewFileModel(true)
        await SaveFile({path: "index.js", bucket: "", body: ""})
    }

    const switchDrawer = async () => {
        setDrawer(!drawer)
    }

    return (
        <div className="App">
            <div className="App-header flex-col h-screen space-y-2 bg-gray-1A1E2A p-2">
                <section className="flex w-full bg-gray-272C38 rounded-lg space-x-2 p-2 items-center leading-none	">
                    <div onClick={switchDrawer}>menu</div>
                    <div>header</div>
                </section>
                <section className="flex-1 flex h-0">
                    <div className="drawer drawer-mobile h-auto">
                        <input type="checkbox" checked={drawer} className="drawer-toggle"/>
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
                                    onFileClick={onFileClick}
                                    onNewFileClick={onNewFileClick}
                                ></FileBrowser>
                            </div>
                        </div>
                    </div>
                </section>
            </div>

            <input type="checkbox" id="my-modal" className="modal-toggle" checked={newFileModel}/>
            <div className="modal">
                <div className="modal-box">
                    <h3 className="font-bold text-lg">Congratulations random Internet user!</h3>
                    <p className="py-4">You've been selected for a chance to get one year of subscription to use
                        Wikipedia for free!</p>
                    <div className="modal-action">
                        <label className="btn" onClick={() => setNewFileModel(false)}>Yay!</label>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default App;
