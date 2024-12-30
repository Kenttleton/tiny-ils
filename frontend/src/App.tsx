//import * as React from 'react';
import { CssVarsProvider } from '@mui/joy/styles';
import CssBaseline from '@mui/joy/CssBaseline';
import Input from '@mui/joy/Input'



function App() {
 // const search = React.useRef<string>('')

  return (
    <CssVarsProvider>
      <CssBaseline />
      <Input
        color="neutral"
        size="lg"
        variant="soft"
        placeholder='What can I help you find?'
      />
    </CssVarsProvider>
  );
}

export default App
