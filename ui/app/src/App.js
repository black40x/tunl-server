import './App.css';
import BrowserWaringPanel from "./components/BrowserWaringPanel";
import Link from '@mui/joy/Link';
import Typography from "@mui/joy/Typography";
import { CssVarsProvider } from '@mui/joy/styles';

function App() {
    return (
        <CssVarsProvider
            defaultMode="system"
            modeStorageKey="identify-system-mode"
        >
            <div className="App">
                <BrowserWaringPanel/>

                <Typography level="body3" sx={{mb: 2}}>
                    Hosted by tunl.online, view project on the <Link href="https://github.com/black40x/tunl-cli" target="_blank">GitHub</Link>
                </Typography>
            </div>
        </CssVarsProvider>
    );
}

export default App;
