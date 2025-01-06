import { component$, useContext } from "@builder.io/qwik";
import { useLocation } from "@builder.io/qwik-city";
import { ThemeContext } from "~/contexts/theme";

export type NavigationLink = {href: string, label: string}
export const TopNavigation = component$((props: {links: NavigationLink[]})=>{
  const {links} = props;
  const loc = useLocation();
  const theme = useContext(ThemeContext);
    return (
      <div class="topnav" id="topnav">
        {links.map((link)=>{return (<a href={link.href} class={loc.url.pathname === link.href? "topnav-link-active" : ""}>{link.label}</a>)})}
        <div class="split">
          <form action="/">
            <input type="text" placeholder="Search.." name="search"/>
            <button type="submit"><i class="fa fa-search"></i></button>
          </form>
          <button onClick$={() => (theme.value = theme.value == 'dark' ? 'light' : 'dark')}>
            {theme.value == 'dark' ? 'lighten' : 'darken'}
          </button>
        </div>
      </div>
    )
})