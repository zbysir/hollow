import Link from "../component/Link";

export default function Header(props) {
    const menus = [
        {href: '/', name: 'Home'},
        {href: '/tags', name: 'Tags'},
        {href: '/about', name: 'About'},
        {href: '/friend', name: 'Friends'},
    ]
    return <div
    >
        {/* copy from https://devdojo.com/tails/v1/app#_ */}
        <section className="w-full bg-white dark:bg-gray-900 border-b border-gray-800">
            <div
                className="container flex flex-col flex-wrap items-center mx-auto md:flex-row max-w-6xl px-5 py-1">
                <div className="relative flex flex-col md:flex-row">
                    <Link href="/"
                          className="flex items-center my-2 md:my-0 font-medium lg:w-auto lg:items-center lg:justify-center ">
                        <span
                            className="mx-auto text-xl font-black leading-none text-gray-900 dark:text-gray-100 select-none"> {props.name}
                            <span className="text-indigo-600"> .</span>
                        </span>
                    </Link>
                    <div className="md:py-3">
                        <div className="md:pl-4 md:ml-8 md:border-l md:dark:border-gray-700 h-full"></div>
                    </div>
                    <nav
                        className="flex space-x-2 flex-wrap items-center text-lg
                        tracking-wide
                        md:border-gray-200
                         font-medium
                        ">
                        {
                            menus.map(i => (
                                <Link href={i.href} className="p-2 transition duration-150 hover:dark:text-gray-200 ">{i.name}</Link>
                            ))
                        }
                    </nav>
                </div>

                <div className="inline-flex items-center space-x-6 lg:justify-end">
                    {/*<a href="#"*/}
                    {/*   className="text-base font-medium leading-6 text-gray-600 whitespace-no-wrap transition duration-150 ease-in-out hover:text-gray-900">*/}
                    {/*  Sign in*/}
                    {/*</a>*/}
                    {/*<Link href="/"*/}
                    {/*   className="inline-flex items-center justify-center px-4 py-2 text-base font-medium leading-6 text-white whitespace-no-wrap bg-indigo-600 border border-transparent rounded-md shadow-sm hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-600">*/}
                    {/*    Sign up*/}
                    {/*</Link>*/}
                </div>
            </div>
        </section>
    </div>
}