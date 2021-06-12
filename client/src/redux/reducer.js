import produce, { enableAllPlugins } from 'immer';

import { TYPES as actions } from './actions';
import { combineReducers } from 'redux';

enableAllPlugins(); // make Immer uber!

const initialState = {
  connected: false,
  username: null,
  channels: {},
  currentChannel: null,
  joinedChannels: new Set(),
};

function getOrCreateChannel(state, channelName, currentUsername) {
  let channel = state.channels[channelName]
  if (!channel) {
    const channel = { name: channelName, users: new Set(), events: [], activity: false };

    if (currentUsername) {
      channel.users.add(currentUsername);
    }

    state.channels[channelName] = channel;
  }

  return channel;
}

function chatReducer(state = initialState, action) {
  switch (action.type) {
    case actions.CLOSED:
    case actions.ERROR:
      return { ...state, connected: false };

    case actions.CHANNELS_LIST:
      return produce(state, draft => {
        for (const channelName of action.channels) {
          getOrCreateChannel(draft, channelName);
        }
      });

    case actions.CHANNEL_JOINED:
      return produce(state, draft => {
        const channel = getOrCreateChannel(draft, action.channel);
        channel.users.add(action.username);
        channel.events.push(action);
        channel.activity = channel.activity || channel.name !== draft.currentChannel;
      });

    case actions.CHANNEL_LEFT:
      return produce(state, draft => {
        const channel = draft.channels[action.channel];
        const { username } = draft;

        channel.users.delete(action.username);
        channel.events.push(action);
        channel.activity = channel.activity || (action.username !== username && channel.name !== draft.currentChannel);
      });

    case actions.CHANNEL_INFO:
      return produce(state, draft => {
        console.log("ACTION:", action);
        console.log("DRAFT.CHANNELS:", draft.channels);
        draft.channels[action.name].users = new Set(action.users);
      });

    case actions.CHANNEL_CHAT:
      return produce(state, draft => {
        const channel = draft.channels[action.channel];
        channel.events.push(action);
        channel.activity = channel.activity || channel.name !== draft.currentChannel;
      });

    case actions.USER_CONNECTED: {
      return produce(state, draft => {
        draft.connected = true;
        draft.username = action.username;
      });
    }

    case actions.USER_JOIN: {
      return produce(state, draft => {
        const channel = getOrCreateChannel(draft, action.channel);
        channel.users.add(action.username);
        draft.joinedChannels.add(action.channel);
      });
    }

    case actions.USER_LEAVE: {
      return produce(state, draft => {
        const channel = draft.channels[action.channel];

        draft.joinedChannels.delete(action.channel);
        if (draft.currentChannel === action.channel) {
          draft.currentChannel = null;
        }
        channel.activity = false;
      });
    }

    case actions.SET_CHANNEL: {
      return produce(state, draft => {
        draft.currentChannel = action.channel;
        draft.channels[action.channel].activity = false;
      });
    }

    default:
      return state;
  }
}

const reducers = combineReducers({
  chat: chatReducer,
});

export default reducers;