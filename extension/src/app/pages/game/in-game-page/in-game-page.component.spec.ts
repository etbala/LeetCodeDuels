import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InGamePageComponent } from './in-game-page.component';

describe('InGamePageComponent', () => {
  let component: InGamePageComponent;
  let fixture: ComponentFixture<InGamePageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InGamePageComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(InGamePageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
