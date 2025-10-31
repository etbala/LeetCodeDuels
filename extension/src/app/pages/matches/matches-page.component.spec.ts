import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MatchesPageComponent } from './matches-page.component';

describe('SettingsPageComponent', () => {
  let component: MatchesPageComponent;
  let fixture: ComponentFixture<MatchesPageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MatchesPageComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MatchesPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
