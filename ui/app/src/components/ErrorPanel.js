import React from 'react';
import Box from '@mui/joy/Box';
import Alert from '@mui/joy/Alert';
import Typography from '@mui/joy/Typography';

const ErrUndefinedClient = 1001
const ErrConnectClient = 1002
const ErrReceiveData = 1003
const ErrClientResponse = 2000

function ErrorPanel(props) {
    let {message} = props

    return (
        <section className="ErrorPanel">
            <Box sx={{display: 'flex', gap: 2, flexDirection: 'column', mb: 3, mr: 1, ml: 1}}>
                <Alert
                    sx={{alignItems: 'flex-start', flexDirection: 'column'}}
                    color="danger"
                    variant="outlined"
                    size="lg"
                >
                    <Typography fontWeight="lg" mt={0.25}>
                        ERROR_{message.code}
                    </Typography>
                    <Typography fontSize="sm" sx={{opacity: 0.8, textAlign: 'left'}}>
                        {(message.code === ErrUndefinedClient) ? (
                            <box>
                                Tunnel not found: {message.tunl_id}
                            </box>
                        ) : ""}
                        {(message.code === ErrConnectClient) ? (
                            <div>
                                Can't connect to the tunnel. <b>REMOTE IP</b>: {message.remote_ip}. <b>TUNL
                                HOST</b>: {message.tunl_host}
                            </div>
                        ) : ""}
                        {(message.code === ErrReceiveData) ? (
                            <box>
                                Can't receive data from the tunnel. <b>REMOTE IP</b>: {message.remote_ip}. <b>TUNL
                                HOST</b>: {message.tunl_host}
                            </box>
                        ) : ""}
                        {(message.code === ErrClientResponse) ? (
                            <box>
                                Error client response. <b>REMOTE IP</b>: {message.remote_ip}. <b>TUNL
                                HOST</b>: {message.tunl_host}
                            </box>
                        ) : ""}
                    </Typography>
                </Alert>
            </Box>
        </section>
    );
}

export default ErrorPanel;
