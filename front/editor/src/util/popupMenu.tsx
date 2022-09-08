import {Menu} from "../component/Menu";
import ReactDOM from "react-dom/client";
import {MenuI} from "../component/Header";

interface showContentMenuParams {
    x: number
    y: number
    menu: MenuI[]
    onClick: (m: MenuI) => void
    id: string
    mask?: boolean
}

export function ShowPopupMenu(p: showContentMenuParams) {
    // 删除上一次的
    const d = document.getElementById('cm:' + p.id)
    if (d) {
        d.remove()
    }

    const dom = document.createElement('div');
    dom.id = 'cm:' + p.id
    if (p.mask) {
        dom.className = "h-full w-full absolute inset-0 z-20"
    }

    const m = <div
        className="bg-gray-272C38 border border-gray-600 rounded-sm text-sm text-white py-1 " style={
        {position: "fixed", left: p.x + "px", top: p.y + "px", minWidth: '140px', zIndex: 1000}
    }>
        <Menu menus={p.menu} onMenuClick={(m) => p.onClick(m)}></Menu>
    </div>

    ReactDOM.createRoot(dom).render(m);

    document.body.appendChild(dom);

    setTimeout(() => {
        document.body.addEventListener('click', (e) => {
            dom.remove()
        }, {once: true})
    })

    return dom
}
