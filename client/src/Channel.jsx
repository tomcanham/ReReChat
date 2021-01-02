import React from 'react';
import { connect } from 'react-redux';
import styled from 'styled-components';
import ChannelEvents from './ChannelEvents';
import ChannelUsers from './ChannelUsers';
import InputBox from './InputBox';
import { leaveChannel } from './redux/actions';

const Chat = styled.div`
  padding: 0.5rem;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
`;

const EventsAndUsersWrapper = styled.div`
  display: flex;
  flex-direction: row;
  flex-grow: 1;
`;

const ChannelHeaderWrapper = styled.div`
  display: flex;
  flex-direction: row;
  user-select: none;
`;

const ChannelTitleWrapper = styled.div`
  flex-grow: 1;
  font-size: 2rem;
  font-weight: bold;
  text-align: center;
`;

const ChannelLeaveButton = styled.div`
  font-size: 1.5rem;
  font-family: sans-serif;
  font-weight: bolder;
  text-align: center;
  cursor: pointer;
  &::after {
    content: "X";
  }
`;

function ChannelHeader({ name, isJoined, onLeave }) {
  return (
    <ChannelHeaderWrapper>
      <ChannelTitleWrapper>{name}</ChannelTitleWrapper>
      {isJoined && <ChannelLeaveButton onClick={() => onLeave(name)} />}
    </ChannelHeaderWrapper>
  );
}

function Channel({ connected, currentChannel, joinedChannels, name, onLeave }) {
  const isJoined = connected && joinedChannels.has(currentChannel);
  const isActive = isJoined && currentChannel === name;

  return (
      <Chat>
        <ChannelHeader name={currentChannel} isJoined={isJoined} onLeave={onLeave} />
        <EventsAndUsersWrapper>
          <ChannelEvents channel={currentChannel} />
          <ChannelUsers channel={currentChannel} />
        </EventsAndUsersWrapper>
        {isActive && <InputBox channel={currentChannel} />}
      </Chat>
  );
}

const mapStateToProps = (state) => ({
  connected: state.chat.connected,
  currentChannel: state.chat.currentChannel,
  joinedChannels: state.chat.joinedChannels,
});

const mapDispatchToProps = (dispatch) => ({
  onLeave: (name) => {
    dispatch(leaveChannel(name));
  }
})

export default connect(mapStateToProps, mapDispatchToProps)(Channel);