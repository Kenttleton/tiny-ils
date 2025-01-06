import { createContextId, Signal } from "@builder.io/qwik";

export const ThemeContext = createContextId<Signal<string>>(
    'docs.theme-context'
  );