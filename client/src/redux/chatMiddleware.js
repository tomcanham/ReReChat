import { TYPES as actions } from './actions';

let sendMessage;

const chatMiddleware = ({ dispatch }) => (next) => (action) => {
  // eslint-disable-next-line default-case
  switch (action.type) {
    case actions.CONNECT: {
      const { url, protocol } = action;
      const ws = new WebSocket(url, protocol);
      sendMessage = null;

      ws.onopen = function() {
        sendMessage = (msg) => ws.send(JSON.stringify(msg));
      }

      ws.onclose = function() {
        sendMessage = null;
        next({ type: actions.CLOSED });
      }

      ws.onmessage = function(msg) {
        const message = JSON.parse(msg.data);
        next({ ...message, type: `chat/server/${message.type}` });
      }

      ws.onerror = function(error) {
        sendMessage = null;
        next({ type: actions.ERROR, error });
      }

      return;
    }

    case actions.SEND_MESSAGE: {
      if (sendMessage) {
        const payload = { ...action };
        delete payload.msgType;
        payload.type = action.msgType;

        sendMessage(payload);
      }
      return;
    }
  }

  return next(action);
};

export default chatMiddleware;