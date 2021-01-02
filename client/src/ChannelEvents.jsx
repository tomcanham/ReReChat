import React from 'react';
import { connect } from 'react-redux';
import styled from 'styled-components';
import { TYPES as actions } from './redux/actions';

const EventsWrapper = styled.div`
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  border: 1px solid black;
  margin-right: 0.5rem;
  padding: 0.25rem;
`;

const ChatMessageWrapper = styled.div`
  display: flex;
  flex-direction: row;
  padding: .1rem;
`;

const ChatSender = styled.span`
  min-width: 100px;
  background-color: silver;
  text-align: center;
  padding: .1rem;
`;

const ChatMessageText = styled.span`
  flex-grow: 1;
  padding: .1rem .1rem .1rem 1rem;
`;

const ChannelEventWrapper = styled.div`
  font-style: italic;
  padding: .1rem .1rem .1rem 1rem;
`;

function ChatMessage({ msg }) {
  return (
    <ChatMessageWrapper>
      <ChatSender>{msg.sender}</ChatSender>
      <ChatMessageText>{msg.message}</ChatMessageText>
    </ChatMessageWrapper>
  );
}

function JoinMessage({ msg }) {
  return (
    <ChannelEventWrapper>
      {msg.username} has joined channel {msg.channel}
    </ChannelEventWrapper>
  );
}

function LeaveMessage({ msg }) {
  return (
    <ChannelEventWrapper>
      {msg.username} has left channel {msg.channel}
    </ChannelEventWrapper>
  );
}

function ChannelEvent({ msg }) {
  switch (msg.type) {
    case actions.CHANNEL_JOINED:
      return <JoinMessage msg={msg} />

    case actions.CHANNEL_LEFT:
      return <LeaveMessage msg={msg} />

    default:
      return <ChatMessage msg={msg} />
  }
}

export function ChannelEvents({ events = [] }) {
  return (
      <EventsWrapper>
        {events.map((msg, idx) => <ChannelEvent msg={msg} key={`event-${idx}`} />)}
      </EventsWrapper>
  );
}

function mapStateToProps(state, ownProps) {
  const { channels, currentChannel } = state.chat;

  return currentChannel ? {
    events: channels[currentChannel]?.events,
  } : {
    events: [],
  };
}

export default connect(mapStateToProps)(ChannelEvents);
