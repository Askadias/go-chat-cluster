import {range} from "rxjs/observable/range";
import {timer} from "rxjs/observable/timer";
import {zip} from "rxjs/observable/zip";
import {environment as env} from "../../environments/environment";

export function isValidURL(str: string) {
  const pattern = new RegExp('^((http|https):\\/)?\\/?' + // protocol
    '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.?)+[a-z]{2,}|' + // domain name
    '((\\d{1,3}\\.){3}\\d{1,3}))' + // OR ip (v4) address
    '(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*' + // port and path
    '(\\?[;&a-z\\d%_.~+=-]*)?' + // query string
    '(\\#[-a-z\\d_]*)?$', 'i'); // fragment locator
  return pattern.test(str);
}

export function exponentialBackOff(errors) {
  return zip(
    range(1, env.socket.maxSubscriptionRetries), errors, function (i, e) {
      return i
    })
    .flatMap(function (i) {
      console.log("delay retry by " + i + " second(s)");
      return timer(i * 1000);
    });
}
