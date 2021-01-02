import React from 'react';
import { connect } from 'react-redux';
import styled from 'styled-components';
import { joinChannel, setChannel } from './redux/actions';

const ChannelsListWrapper = styled.div`
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgb(0, 0, 0);
  padding: 0.25rem;
  width: 200px;
`;

const ChannelNameWrapper = styled.div`
  background-color: ${props => props.isCurrent ? 'yellow' : 'white'};
  cursor: pointer;
  user-select: none;
  font-weight: ${props => props.activity ? 'bold' : 'normal'};
`;

const noOp = () => {};

function ChannelName({ channel, channelName, canJoin, isCurrent, onSelectedChannelChanged }) {
  return (
    <ChannelNameWrapper
      activity={channel.activity}
      canJoin={canJoin}
      isCurrent={isCurrent}
      onClick={isCurrent ? noOp : onSelectedChannelChanged}>{channelName}</ChannelNameWrapper>
  );
}

export function ChannelsList({ channels, currentChannel, joinedChannels, onSelectedChannelChanged }) {
  const channelNames = Object.getOwnPropertyNames(channels).sort();

  return (
    <ChannelsListWrapper>
      {channelNames?.map((name, idx) => (
      <ChannelName
        channel={channels[name]}
        canJoin={!joinedChannels.has(name)}
        channelName={name}
        isCurrent={name === currentChannel}
        onSelectedChannelChanged={() => onSelectedChannelChanged(name, !joinedChannels.has(name))}
        key={`channel-${idx}`} />
        )
      )}
    </ChannelsListWrapper>
  );
}

function mapStateToProps(state) {
  const { channels, currentChannel, joinedChannels } = state.chat;

  return {
    channels,
    currentChannel,
    joinedChannels,
  };
}

function mapDispatchToProps(dispatch) {
  return {
    onSelectedChannelChanged: (channelName, canJoin) => {
      canJoin && dispatch(joinChannel(channelName));
      dispatch(setChannel(channelName));
    },
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ChannelsList);