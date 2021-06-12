import { TYPES as actions } from './actions';

let sendMessage;

function sendMessageInternal(_ws, msgType, msg) {
  console.log('Sending:', msgType, msg);
  _ws.send(`${msgType}\n${JSON.stringify(msg)}`);
}

const chatMiddleware = ({ dispatch }) => (next) => (action) => {
  // eslint-disable-next-line default-case
  switch (action.type) {
    case actions.CONNECT: {
      const { url, protocol } = action;
      const ws = new WebSocket(url, protocol);
      sendMessage = null;

      ws.onopen = function() {
        sendMessage = (msgType, msg) => sendMessageInternal(ws, msgType, msg);
        next({ type: actions.USER_CONNECTED });
      }

      ws.onclose = function(reason) {
        console.error('WebSocket closed:', reason);
        sendMessage = null;
        next({ type: actions.CLOSED });
      }

      ws.onmessage = function(msg) {
        console.log(msg);
        const [msgType, payload] = msg.data.split("\n", 2);

        const message = JSON.parse(payload);
        next({ ...message, type: `chat/server/${msgType}` });
      }

      ws.onerror = function(error) {
        console.error('WebSocket error:', error);
        sendMessage = null;
        next({ type: actions.ERROR, error });
      }

      return;
    }

    case actions.SEND_MESSAGE: {
      if (sendMessage) {
        const payload = { ...action };
        delete payload.msgType;

        sendMessage(action.msgType, payload);
      }
      return;
    }
  }

  return next(action);
};

export default chatMiddleware;