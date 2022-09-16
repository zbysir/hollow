import Container from "../component/Container";

interface FriendLink {
    url: string
    name: string
    info?: string
}

interface Props {
    links: FriendLink[]
}

import hollow from "@bysir/hollow"
let params = hollow.getConfig();
let gallery = params.gallery

export default function Gallery() {
    return <Container>
        <h3>我的拍照、摄影，不过还在搭建中</h3>
        <div className="grid grid-cols-4 gap-8
        auto-rows-auto	auto-cols-auto	grid-flow-row
        ">
            {
                gallery?.map(i => (
                    <img className={"aspect-square object-cover"} src={i.img} alt={i.img}/>
                ))
            }
        </div>
        {/*<div className="prose dark:prose-invert">*/}
        {/*    <div></div>*/}
        {/*</div>*/}
    </Container>

}