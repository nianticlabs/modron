import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { NotificationStore } from '../state/notification.store';

@Component({
  selector: 'app-notification-exceptions',
  templateUrl: './notification-exceptions.component.html',
  styleUrls: ['./notification-exceptions.component.scss'],
})
export class NotificationExceptionsComponent implements OnInit {
  displayedColumns = [
    'userEmail',
    'notificationName',
    'justification',
    'sourceSystem',
    'validUntilTime',
    '$actions',
  ];
  searchText: string;

  constructor(route: ActivatedRoute, public store: NotificationStore) {
    this.searchText = route.snapshot.paramMap.get('notificationName') ?? '';
  }

  ngOnInit(): void {}
}
