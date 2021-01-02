import jwt from 'jsonwebtoken';
import React, { useEffect } from 'react';
import { connect } from 'react-redux';
import styled from 'styled-components';
import Channel from './Channel';
import ChannelsList from './ChannelsList';
import { getChannelsList as _getChannelsList, doConnect } from './redux/actions';

const AppWrapper = styled.div`
  display: flex;
  flex-direction: row;
  height: 100vh;
`;

const getUserName = () => 'tom' + Math.floor(Math.random() * 1000);

function getJwt(userName) {
  const payload = {
    userName,
    iat: Math.floor(Date.now() / 1000) - 30, // 10 minutes ago
  };

  return jwt.sign(payload, 'secret', { expiresIn: '1h'});
}

const userName = getUserName();
const token = getJwt(userName);

export function App({ doConnect, connected, getChannelsList, currentChannel }) {
  // todo: Redux-ize this fake "login" and connect
  // todo: handle disconnects and failures to connect gracefully
  useEffect(() => {
    doConnect();
  }, [doConnect]);

  useEffect(() => {
    if (connected) {
      getChannelsList();
    }
  }, [connected, getChannelsList]);

  return (
    <AppWrapper>
      <ChannelsList />
      <Channel name={currentChannel} />
    </AppWrapper>
  );
}

const mapStateToProps = (state) => ({
  connected: state.chat.connected,
  currentChannel: state.chat.currentChannel,
});

const mapDispatchToProps = (dispatch, ownProps) => ({
  doConnect: () => dispatch(doConnect({ url: `ws://localhost:8080/ws?jwt=${token.toString()}`, protocol: token.toString() })),
  getChannelsList: () => dispatch(_getChannelsList()),
});

export default connect(mapStateToProps, mapDispatchToProps)(App);
