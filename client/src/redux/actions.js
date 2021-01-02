// action types prefixed with 'client' originate from or are purely
// involved with the client (UI, etc.). Action types prefixed with
// 'server' originate on the server and indicate activity or responses
// to commands.
export const TYPES = {
  CONNECT: 'chat/client/connect',
  CLOSED: 'chat/server/closed',
  ERROR: 'chat/server/error',
  SEND_MESSAGE: 'chat/client/send_message',
  CHANNEL_JOINED: 'chat/server/channel.joined',
  CHANNEL_LEFT: 'chat/server/channel.left',
  CHANNEL_INFO: 'chat/server/channel.info',
  CHANNEL_CHAT: 'chat/server/channel.chat',
  USER_CONNECTED: 'chat/server/user.connected',
  USER_JOIN: 'chat/server/user.join',
  USER_LEAVE: 'chat/server/user.leave',
  SET_CHANNEL: 'chat/client/set_channel',
  CHANNELS_LIST: 'chat/server/channels.list'
};

export function setChannel(channel) {
  return {
    type: TYPES.SET_CHANNEL,
    channel,
  };
}

export function getChannelsList() {
  return sendMessage({ type: 'channels.list' });
}

export function doConnect({ url, protocol }) {
  return {
    type: TYPES.CONNECT,
    url,
    protocol,
  };
}

export function sendMessage(msg) {
  const action = { ...msg };
  action.type = TYPES.SEND_MESSAGE;
  action.msgType = msg.type;

  return action;
}

export function sendChatMessage(channel, text) {
  return sendMessage({ type: 'channel.chat', channel, message: text });
}

export function joinChannel(channel) {
  return sendMessage({ channel, type: 'channel.join' });
}

export function leaveChannel(channel) {
  return sendMessage({ channel, type: 'channel.leave' });
}