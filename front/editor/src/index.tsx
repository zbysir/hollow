import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import {Helmet} from "react-helmet";

const root = ReactDOM.createRoot(
    document.getElementById('root') as HTMLElement
);
root.render(
    <React.StrictMode>
        <App />

        <Helmet>
            <meta charSet="utf-8" />
            <title>My Title</title>
            <meta name="apple-mobile-web-app-capable" content="yes" />
            <meta name="x5-orientation" content="portrait"/>
            <meta name="full-screen" content="yes"/>


        </Helmet>
    </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
