package narration

import (
	"fmt"
)

// HPSeverity represents the severity level of HP change
type HPSeverity int

const (
	HPSeverityMinor    HPSeverity = iota // -1 to -10
	HPSeverityModerate                   // -11 to -30
	HPSeverityMajor                      // -31 to -50
	HPSeverityLethal                     // <= -50
)

// SANSeverity represents the severity level of SAN change
type SANSeverity int

const (
	SANSeverityMinor    SANSeverity = iota // -1 to -10
	SANSeverityModerate                    // -11 to -30
	SANSeverityMajor                       // -31 to -50
	SANSeverityLethal                      // <= -50
)

// DescribeHPChange generates a narrative description for HP changes
//
// AC #6: HP 變化描述（80-120 字）
//
// Parameters:
//   - delta: HP change value (negative = damage, positive = healing)
//   - reason: The reason for the HP change (from JudgeResult)
//
// Returns:
//   - string: Narrative description of the HP change
func DescribeHPChange(delta int, reason string) string {
	if delta > 0 {
		// HP 回復
		return describeHPRecovery(delta, reason)
	} else if delta < 0 {
		// HP 下降
		return describeHPDamage(delta, reason)
	}
	// delta == 0, no change
	return ""
}

// describeHPDamage describes HP damage
func describeHPDamage(delta int, reason string) string {
	severity := getHPSeverity(delta)

	var baseDesc string
	switch severity {
	case HPSeverityMinor:
		baseDesc = "你感到一陣疼痛"
	case HPSeverityModerate:
		baseDesc = "你感到一陣劇痛，身體搖搖欲墜"
	case HPSeverityMajor:
		baseDesc = "你遭受了重創，意識開始模糊"
	case HPSeverityLethal:
		baseDesc = "致命傷害讓你失去知覺"
	default:
		baseDesc = "你受到了傷害"
	}

	if reason != "" {
		return fmt.Sprintf("%s。%s（HP %d）", baseDesc, reason, delta)
	}
	return fmt.Sprintf("%s（HP %d）", baseDesc, delta)
}

// describeHPRecovery describes HP recovery
func describeHPRecovery(delta int, reason string) string {
	var baseDesc string
	if delta <= 10 {
		baseDesc = "你感覺好了一些"
	} else if delta <= 30 {
		baseDesc = "你感覺好多了，體力逐漸恢復"
	} else {
		baseDesc = "你感覺煥然一新，傷勢大幅好轉"
	}

	if reason != "" {
		return fmt.Sprintf("%s。%s（HP +%d）", baseDesc, reason, delta)
	}
	return fmt.Sprintf("%s（HP +%d）", baseDesc, delta)
}

// getHPSeverity determines the severity level of HP damage
func getHPSeverity(delta int) HPSeverity {
	// delta is negative for damage
	if delta >= -10 {
		return HPSeverityMinor
	} else if delta >= -30 {
		return HPSeverityModerate
	} else if delta >= -50 {
		return HPSeverityMajor
	}
	return HPSeverityLethal
}

// DescribeSANChange generates a narrative description for SAN changes
//
// AC #6: SAN 變化描述（80-120 字）
//
// Parameters:
//   - delta: SAN change value (negative = sanity loss, positive = recovery)
//   - reason: The reason for the SAN change (from JudgeResult)
//
// Returns:
//   - string: Narrative description of the SAN change
func DescribeSANChange(delta int, reason string) string {
	if delta > 0 {
		// SAN 回復
		return describeSANRecovery(delta, reason)
	} else if delta < 0 {
		// SAN 下降
		return describeSANLoss(delta, reason)
	}
	// delta == 0, no change
	return ""
}

// describeSANLoss describes SAN loss
func describeSANLoss(delta int, reason string) string {
	severity := getSANSeverity(delta)

	var baseDesc string
	switch severity {
	case SANSeverityMinor:
		baseDesc = "你感到一絲不安"
	case SANSeverityModerate:
		baseDesc = "你的理智開始動搖，周圍的一切變得不太真實"
	case SANSeverityMajor:
		baseDesc = "你的理智搖搖欲墜，現實變得扭曲而詭異"
	case SANSeverityLethal:
		baseDesc = "你的意識崩潰了，瘋狂吞噬了你的思維"
	default:
		baseDesc = "你的理智受到衝擊"
	}

	if reason != "" {
		return fmt.Sprintf("%s。%s（SAN %d）", baseDesc, reason, delta)
	}
	return fmt.Sprintf("%s（SAN %d）", baseDesc, delta)
}

// describeSANRecovery describes SAN recovery
func describeSANRecovery(delta int, reason string) string {
	var baseDesc string
	if delta <= 10 {
		baseDesc = "你的心情稍微平靜下來"
	} else if delta <= 30 {
		baseDesc = "你的心情平靜下來，理智逐漸恢復"
	} else {
		baseDesc = "你感到前所未有的平靜，理智完全恢復"
	}

	if reason != "" {
		return fmt.Sprintf("%s。%s（SAN +%d）", baseDesc, reason, delta)
	}
	return fmt.Sprintf("%s（SAN +%d）", baseDesc, delta)
}

// getSANSeverity determines the severity level of SAN loss
func getSANSeverity(delta int) SANSeverity {
	// delta is negative for loss
	if delta >= -10 {
		return SANSeverityMinor
	} else if delta >= -30 {
		return SANSeverityModerate
	} else if delta >= -50 {
		return SANSeverityMajor
	}
	return SANSeverityLethal
}
