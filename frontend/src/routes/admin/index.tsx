import { component$ } from "@builder.io/qwik";
import { useLocation, type DocumentHead } from "@builder.io/qwik-city";

export default component$(() => {
  const loc = useLocation()
  return (
    <>
      <h1>Hello Tiny Librarian!</h1>
    </>
  );
});

export const head: DocumentHead = {
  title: "Tiny Librarian",
  meta: [
    {
      name: "description",
      content: "",
    },
  ],
};