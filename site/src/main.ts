import "./style.css";
import { mountApp } from "./ui/app";

const appEl = document.querySelector<HTMLDivElement>("#app");
if (!appEl) throw new Error("Missing #app element");

mountApp(appEl);

