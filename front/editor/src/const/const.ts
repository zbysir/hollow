let dev = window.location.port === "3001" || window.location.port === "3000";
export let serviceAddress = dev ? "//localhost:9432" : ('//' + window.location.host)