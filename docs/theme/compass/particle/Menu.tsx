import hollow, {Article, getArticles} from "@bysir/hollow"
import {sortBlog} from "../utilx";

interface Menu extends Article {
    link: string
    children?: any
}

interface Props {
    activityMenu: Menu
    menu: Menu[]
}

export default function Menu(props: Props) {
    return <div
    >
        {props.menu.map(i => {
            let name = i.meta?.title || i.name

            return <div>
                    {
                        i.is_dir ?
                            <div className="text-gray-400 my-4">{name}</div> :
                            <div className={`${i.link === props.activityMenu.link ? 'text-red-700' : ''}`}>
                                <a href={i.link}>{name}</a>
                            </div>
                    }

                    <div className="pl-0">
                        <Menu activityMenu={props.activityMenu} menu={i.children}></Menu>
                    </div>

                </div>
            }
        )}

    </div>
}