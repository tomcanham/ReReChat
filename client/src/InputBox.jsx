import React, { useState } from 'react';
import { connect } from 'react-redux';
import styled from 'styled-components';
import { sendChatMessage } from './redux/actions';

const InputWrapper = styled.div`
  margin-top: 0.5rem;
  border-style: 1px solid silver;
  display: flex; 
  flex-direction: row;
`;

const InputText = styled.input`
  flex-grow: 1;
  margin-right: 0.5rem;
`;

const InputButton = styled.div`
  color: ${props => props.disabled ? 'grey' : 'black'};
  cursor: ${props => props.disabled ? 'default' : 'pointer'};
  border: 1px solid black;
  border-radius: 0.25rem;
  background-color: ${props => props.disabled ? 'silver' : 'white'};
  width: 200px;
  padding: 0.25rem;
  text-align: center;
`;

// chatMessage := &ChatMessage{Type: "channel.chat", Channel: channelName.(string), Sender: c.username, Message: text.(string)}
export const Input = ({ sendMessage }) => {
  const [text, setText] = useState('');

  const send = () => {
    if (text.length > 0) {
      sendMessage(text);
      setText(() => '');
    }
  }

  return (
    <InputWrapper>
      <InputText
        type="text"
        value={text}
        onChange={(e) => setText(() => e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && send()} />
      <InputButton disabled={text.length === 0} onClick={send}>Send</InputButton>
    </InputWrapper>
  );
}

const mapDispatchToProps = (dispatch, ownProps) => ({
  sendMessage: (msg) => dispatch(sendChatMessage(ownProps.channel, msg)),
});

export default connect(null, mapDispatchToProps)(Input);