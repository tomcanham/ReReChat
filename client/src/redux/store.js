import { applyMiddleware, compose, createStore } from 'redux';
import chatMiddleware from './chatMiddleware';
import rootReducer from './reducer';

const composeEnhancers = window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__ || compose;
const middlewareEnhancer = applyMiddleware(chatMiddleware);
const store = createStore(rootReducer, /* preloadedState, */ composeEnhancers(middlewareEnhancer));

export default store;