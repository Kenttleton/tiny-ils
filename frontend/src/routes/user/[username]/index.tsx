import { component$ } from "@builder.io/qwik";
import { DocumentHead, useLocation } from "@builder.io/qwik-city";

export default component$(() => {
    const loc = useLocation();
  return (
    <>
      <h1>Hi {loc.params.username}👋</h1>
    </>
  );
});

export const head: DocumentHead = {
  title: "Tiny Patron",
  meta: [
    {
      name: "description",
      content: "",
    },
  ],
};