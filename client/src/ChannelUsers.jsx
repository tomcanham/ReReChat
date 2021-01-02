import React from 'react';
import { connect } from 'react-redux';
import styled from 'styled-components';

const ChannelUsersWrapper = styled.div`
  display: flex;
  flex-direction: column;
  border: 1px solid rgb(0, 0, 0);
  padding: 0.25rem;
  width: 200px;
`;

const ChannelUserWrapper = styled.div`
  text-decoration: ${props => props.isCurrent ? 'underline' : 'none'};
`;

function ChannelUser({ username, isCurrent }) {
  return (
    <ChannelUserWrapper isCurrent={isCurrent}>{username}</ChannelUserWrapper>
  );
}

export function ChannelUsers({ username, users }) {
  return (
    <ChannelUsersWrapper>
      {users.map((user, idx) => <ChannelUser isCurrent={username === user} username={user} key={`user-${idx}`} />)}
    </ChannelUsersWrapper>
  );
}

function mapStateToProps(state, ownProps) {
  const { channels, username } = state.chat;
  const channel = channels[ownProps.channel];

  return {
    username,
    users: [...channel?.users || []],
  };
}

export default connect(mapStateToProps)(ChannelUsers);