import Logo from "./logo";

export let meta = {}
export default function () {
    return <div>
        <h1 className="bg-base-400 text-center">
            Hollow
        </h1>
        <div className="flex">
            <p className={"flex-1"}>1</p>
            <Logo></Logo>
            <p>2</p>
        </div>
    </div>
}