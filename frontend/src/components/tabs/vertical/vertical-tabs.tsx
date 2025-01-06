import { component$, JSXChildren } from "@builder.io/qwik"

export type TabValue = { id: string, slot: JSXChildren}
export type TabKey = {id: string, label: string, onclick: ()=>{}}
export const VerticalTabs = component$((tabs: Map<TabKey,TabValue>) => {
    const keys = Array.from(tabs.keys())
    const values = Array.from(tabs.values())
    return (
        <div class="tabs-vertical">
            <div class="tab">
                {keys.map((key)=> {
                    return (<button class="tablinks" id={key.id} onclick$={key.onclick}>{key.label}</button>)
                })}
            </div>
          {values.map((value)=>{
            return ( 
                <div id={value.id} class="tabcontent">
                    {value.slot}
                </div>
            )
          })}
        </div>
    )
})