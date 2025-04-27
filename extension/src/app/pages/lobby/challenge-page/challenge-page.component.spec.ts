import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChallengePageComponent } from './challenge-page.component';

describe('ChallengePageComponent', () => {
  let component: ChallengePageComponent;
  let fixture: ComponentFixture<ChallengePageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ChallengePageComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ChallengePageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
