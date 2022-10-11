import { Injectable } from '@angular/core';
import { User, AuthenticationService } from '../authentication.service';

@Injectable()
export class AuthenticationStore {
  private _user = new User(false);

  constructor(private _service: AuthenticationService) {
    this.fetchInitialData();
  }

  get user(): User {
    return this._user;
  }

  private fetchInitialData() {
    this._user = this._service.authenticate();
  }
}
