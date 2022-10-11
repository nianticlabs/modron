import { Injectable } from '@angular/core';
import { CookieService } from 'ngx-cookie-service';

export class User {
  constructor(private _signedIn: boolean, private _email = '') {}

  get email(): string {
    return this._email;
  }

  get signedIn(): boolean {
    return this._signedIn;
  }
}

@Injectable({
  providedIn: 'root',
})
export class AuthenticationService {
  public static readonly USER_EMAIL_COOKIE_NAME = 'modron-user-email';

  constructor(private _service: CookieService) {}

  authenticate(): User {
    const email = this._service.get(
      AuthenticationService.USER_EMAIL_COOKIE_NAME
    );
    if (email !== '') {
      return new User(true, email);
    }
    return new User(false);
  }
}
