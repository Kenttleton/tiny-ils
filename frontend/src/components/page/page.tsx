import { component$, Slot } from "@builder.io/qwik";
import { TopNavigation } from "../top-navigation/top-navigation";
import {Body} from "~/components/body/body"

const links = [{href: "/", label: "Home"},{href: "/admin", label: "Admin"}]
export const Page = component$(()=>{
    return (<><TopNavigation links={links}/><Body><Slot/></Body></>)
})