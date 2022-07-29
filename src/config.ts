import Index from "./Index"

// @ts-ignore
import db from "db"

let blog = db.getBlog('./blogs');

let global = {
  title: "bysir 的博客",
  me: "bysir",
}

export default {
  pages: [
    {
      name: 'index',
      component: () => Index({...global, page: 'home', page_data: {blogs: blog}}),
    },
    ...blog.map(b => ({
      name: 'blogs/' + b.name,
      component: () => {
        b.content = b.getContent()
        return Index({...global, page: 'blog-detail', page_data: b})
      }
    }))
  ],
  assert: ['tailwind.css']
}