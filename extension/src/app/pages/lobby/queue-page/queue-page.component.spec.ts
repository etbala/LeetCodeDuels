import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QueuePageComponent } from './queue-page.component';

describe('QueuePageComponent', () => {
  let component: QueuePageComponent;
  let fixture: ComponentFixture<QueuePageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QueuePageComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QueuePageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
