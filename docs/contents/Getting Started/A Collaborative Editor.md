---
title: A Collaborative Editor
desc: A five minute guide to make an editor collaborative
slug: a-collaboration-editor
sort: 1
---

Yjs is a modular framework for syncing things in real-time - like editors!
This guide will walk you through the main concepts of Yjs. First, we are going to create a collaborative editor and sync it with clients. You will get introduced to Yjs documents and to providers, that allow you to sync through different network protocols. Next, we talk about Awareness & Presence which are very important aspects of collaborative software. I created a separate section for Offline Support that shows you how to create offline-ready applications by just adding a few lines of code. The last section is an in-depth guide to Shared Types.

> If you are impatient jump to the live demo at the bottom of the page 😉

Let's get started by deciding on an editor to use. Yjs doesn't ship with a customized editor. There are already a lot of awesome open-source editors projects out there. Yjs supports many of them using extensions. Editor bindings are a concept in Yjs that allow us to bind the state of a third-party editor to a syncable Yjs document. This is a list of all known editor bindings:

For the purpose of this guide, we are going to use the Quill editor - a great rich-text editor that is easy to setup. For a complete reference on how to setup Quill I refer to their documentation. If you first require a basic introduction in npm and bundles, please refer to the webpack getting started guide and additionally setting up a development server.

```js
import Quill from 'quill'
import QuillCursors from 'quill-cursors'

Quill.register('modules/cursors', QuillCursors);

const quill = new Quill(document.querySelector('#editor'), {
  modules: {
    cursors: true,
    toolbar: [
      // adding some basic Quill content features
      [{ header: [1, 2, false] }],
      ['bold', 'italic', 'underline'],
      ['image', 'code-block']
    ],
    history: {
      // Local undo shouldn't undo changes
      // from remote users
      userOnly: true
    }
  },
  placeholder: 'Start collaborating...',
  theme: 'snow' // 'bubble' is also great
})
```

Next, we are going to install Yjs and the y-quill editor binding.

The ytext object is a shared data structure for representing text. It also supports formatting attributes (i.e. bold and italic). Yjs automatically resolves concurrent changes on shared data so we don't have to worry about conflict resolution anymore. Then we synchronize ytext with the quill editor and keep them in-sync using the QuillBinding. Almost all editor bindings work like this. You can simply exchange the editor binding if you switch to another editor.
But don't stop here, the editor doesn't sync to other clients yet! We need to choose a provider or implement our own communication protocol to exchange document updates with other peers.

Each provider has pros and cons. The y-webrtc provider connects clients directly with each other and is a perfect choice for demo applications because it doesn't require you to set up a server. But for a real-world application, you often want to sync the document to a server. In any case, we got you covered. It is easy to change the provider because they all implement the same interface.

> Providers are meshable. You can connect multiple providers to a Yjs instance at the same time. Document updates will automatically sync through the different communication channels. Meshing providers can improve reliability through redundancy and decrease network delay.

But for now, let's enjoy what we built. I included the same fiddle twice so you can observe the editors sync in real-time. Aware, the editor content is synced with all users visiting this page!
