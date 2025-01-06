import { component$, Slot } from "@builder.io/qwik";
import type { RequestHandler } from "@builder.io/qwik-city";
import { Session } from "~/models/session";

export const onRequest: RequestHandler = (event) => {
    const session: Session | null = event.sharedMap.get('session');
    if (!session || new Date(session.expires) < new Date()) {
      throw event.redirect(302, `/`);
    }
  };

export default component$(() => {
  return <Slot />;
});