import React from 'react';
import Cookies from 'js-cookie';
import Button from '@mui/joy/Button';
import Sheet from '@mui/joy/Sheet';
import Typography from '@mui/joy/Typography';

function BrowserWaringPanel(props) {
    let { message } = props

    const visitSite = () => {
        Cookies.set("tunl-online-skip-warning", "1");
        window.location.reload();
    }
    const closeSite = () => {
        Cookies.remove("tunl-online-skip-warning");
        window.close();
    }

    return (
        <section className="BrowserWaringPanel">
            <Sheet
                sx={{
                    maxWidth: 450,
                    mx: 'auto',
                    my: 4,
                    mr: 1,
                    ml: 1,
                    py: 3,
                    px: 2,
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 2,
                    borderRadius: 'sm',
                    boxShadow: 'md',
                }}
                variant="outlined"
            >
                <Typography level="h3">Warning!</Typography>
                <Typography
                    sx={{
                        textAlign: 'left'
                    }}
                    level="body1">
                    Be careful and don't input your payment or personal data on this site, this website may be used as fishing or a hacker attack. If you dont trust this site owner, please close it!
                </Typography>

                <Typography
                    sx={{
                        textAlign: 'left'
                    }}
                    level="body2">
                    <b>REMOTE IP</b>: { message.remote_ip } <br/>
                    <b>TUNL HOST</b>: { message.tunl_host }
                </Typography>

                <Typography
                    sx={{
                        textAlign: 'left'
                    }}
                    level="body2">

                </Typography>

                <Sheet sx={{
                    display: 'flex',
                    flexDirection: 'row'
                }}>
                    <Button sx={{ mr: 1, flex: 1 }}
                            variant="outlined"
                            color="danger"
                            onClick={closeSite}>Close</Button>
                    <Button sx={{ flex: 1 }}
                            variant="solid"
                            onClick={visitSite}>Trust</Button>
                </Sheet>
            </Sheet>
        </section>
    );
}

export default BrowserWaringPanel;
