import { Component, OnInit } from '@angular/core';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-modron-app',
  templateUrl: './modron-app.component.html',
  styleUrls: ['./modron-app.component.scss'],
})
export class ModronAppComponent implements OnInit {
  public organization: string;

  constructor() {
    this.organization = environment.organization;
  }

  ngOnInit(): void {}
}
